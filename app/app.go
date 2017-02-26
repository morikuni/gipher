package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-yaml/yaml"
	"github.com/morikuni/accessor"
	"github.com/morikuni/gipher"
	"github.com/spf13/pflag"
)

type App interface {
	Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int
}

func NewApp() App {
	return app{}
}

type app struct{}

func (a app) Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	flag := pflag.NewFlagSet("gipher", pflag.ContinueOnError)
	help := flag.BoolP("help", "h", false, "print this message.")
	inputFile := flag.StringP("file", "f", "", "file path to input.")
	outputFile := flag.StringP("output", "o", "", "file path to output.")
	format := flag.String("format", "text", `"text", "json", "yaml", or "toml"`)
	field := flag.String("field", "", `field to be encrypted/decrypted (e.g. "user/items"). all fields are encrypted/decrypted by default.`)
	cryptorType := flag.String("cryptor", "password", `"password" or "aws-kms".`)
	awsKeyID := flag.String("aws-key-id", "", "key id for aws kms. (required when cryptor is aws-kms)")
	awsRegion := flag.String("aws-region", "", "aws region. (required when cryptor is aws-kms)")
	dryrun := flag.Bool("dryrun", false, `display fields to be affected as "THIS FIELD WILL BE CHENGED", without operation.`)
	flag.Usage = func() {
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Usage: gipher <command> [flags]")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Commands:")
		fmt.Fprintln(stderr, "      encrypt               encrypt a file.")
		fmt.Fprintln(stderr, "      decrypt               decrypt a encrypted file.")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Flags:")
		fmt.Fprintln(stderr, flag.FlagUsages())
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Environment variables:")
		fmt.Fprintln(stderr, "      GIPHER_PASSWORD       set password without prompt.")
		fmt.Fprintln(stderr, "      AWS_PROFILE           set profile for aws.")
		fmt.Fprintln(stderr, "      AWS_ACCESS_KEY_ID     set access key id for aws.")
		fmt.Fprintln(stderr, "      AWS_SECRET_ACCESS_KEY set secret access key for aws.")
	}

	err := flag.Parse(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if *help {
		flag.Usage()
		return 0
	}

	command := flag.Arg(1)
	if command == "" {
		fmt.Fprintln(stderr, "command is required")
		flag.Usage()
		return 1
	}

	input, output, err := createIO(stdin, stdout, *inputFile, *outputFile)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer input.Close()
	defer output.Close()

	acc, err := decodeToAccessor(*format, input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	target, err := extractTargetByField(acc, *field)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	cryptor, err := createCryptor(*cryptorType, *awsRegion, *awsKeyID)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	err = target.Foreach(func(path accessor.Path, value interface{}) error {
		if *dryrun {
			return target.Set(path, "THIS FIELD WILL BE CHENGED")
		}

		switch command {
		case "encrypt":
			cipher, shouldSet, err := encrypt(cryptor, value)
			if err != nil {
				return err
			}
			if shouldSet {
				return target.Set(path, cipher)
			}
		case "decrypt":
			if s, ok := value.(string); ok {
				value, err := decrypt(cryptor, s)
				if err != nil {
					return err
				}
				return target.Set(path, value)
			}
		default:
			return fmt.Errorf("unknown command: %s", command)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	err = encodeAccessor(*format, output, acc)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}

func decodeToAccessor(format string, input io.Reader) (accessor.Accessor, error) {
	switch format {
	case "":
		return nil, errors.New("type is required")
	case "json":
		var obj interface{}
		err := json.NewDecoder(input).Decode(&obj)
		if err != nil {
			return nil, err
		}
		return accessor.NewAccessor(obj)
	case "yaml":
		bs, err := ioutil.ReadAll(input)
		if err != nil {
			return nil, err
		}
		var obj interface{}
		err = yaml.Unmarshal(bs, &obj)
		if err != nil {
			return nil, err
		}
		return accessor.NewAccessor(obj)
	case "toml":
		var obj interface{}
		_, err := toml.DecodeReader(input, &obj)
		if err != nil {
			return nil, err
		}
		return accessor.NewAccessor(obj)
	case "text":
		bs, err := ioutil.ReadAll(input)
		if err != nil {
			return nil, err
		}
		return accessor.NewAccessor(string(bs))
	default:
		return nil, fmt.Errorf("unknown type: %q", format)
	}
}

func encodeAccessor(format string, output io.Writer, acc accessor.Accessor) error {
	switch format {
	case "":
		return errors.New("type is required")
	case "json":
		return json.NewEncoder(output).Encode(acc.Unwrap())
	case "yaml":
		bs, err := yaml.Marshal(acc.Unwrap())
		if err != nil {
			return err
		}
		_, err = output.Write(bs)
		return err
	case "toml":
		return toml.NewEncoder(output).Encode(acc.Unwrap())
	case "text":
		_, err := output.Write([]byte(acc.Unwrap().(string)))
		return err
	default:
		return fmt.Errorf("unknown type: %q", format)
	}
}

func extractTargetByField(acc accessor.Accessor, field string) (accessor.Accessor, error) {
	if field == "" {
		return acc, nil
	}
	path, err := accessor.ParsePath(field)
	if err != nil {
		return nil, fmt.Errorf("field is invalid: %s", err)
	}
	return acc.Get(path)
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func newNopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

func createIO(stdin io.Reader, stdout io.Writer, inputFile, outputFile string) (io.ReadCloser, io.WriteCloser, error) {
	input := ioutil.NopCloser(stdin)
	if inputFile != "" {
		f, err := os.Open(inputFile)
		if err != nil {
			return nil, nil, err
		}
		input = f
	}

	output := newNopCloser(stdout)
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			input.Close()
			return nil, nil, err
		}
		output = f
	}

	return input, output, nil
}

func createCryptor(cryptor, awsRegion, awsKeyID string) (gipher.Cryptor, error) {
	switch cryptor {
	case "":
		return nil, errors.New("cryptor is required")
	case "password":
		return gipher.NewPasswordCryptorWithPrompt()
	case "aws-kms":
		if awsRegion == "" {
			return nil, errors.New("aws-region is required for aws-kms")
		}
		if awsKeyID == "" {
			return nil, errors.New("key-id is required for aws-kms")
		}
		return gipher.NewAWSKMSCryptor(awsRegion, awsKeyID)
	default:
		return nil, fmt.Errorf("unknown cryptor: %q", cryptor)
	}
}

func decrypt(cryptor gipher.Cryptor, value string) (interface{}, error) {
	text, err := cryptor.Decrypt(gipher.Base64String(value))
	if err != nil {
		return "", err
	}
	return decodeFromString(text)
}

func encrypt(cryptor gipher.Cryptor, value interface{}) (string, bool, error) {
	text, shouldSet, err := encodeToString(value)
	if !shouldSet || err != nil {
		return "", shouldSet, err
	}
	cipher, err := cryptor.Encrypt(text)
	return string(cipher), true, err
}

func encodeToString(value interface{}) (string, bool, error) {
	switch t := value.(type) {
	case string:
		return String + ":" + t, true, nil
	case int:
		return Int + ":" + strconv.FormatInt(int64(t), 10), true, nil
	case int64:
		return Int + ":" + strconv.FormatInt(t, 10), true, nil
	case float32:
		return Float + ":" + strconv.FormatFloat(float64(t), 'E', -1, 64), true, nil
	case float64:
		return Float + ":" + strconv.FormatFloat(t, 'E', -1, 64), true, nil
	default:
		return "", false, nil
	}
}

func decodeFromString(text string) (interface{}, error) {
	s := strings.SplitN(text, ":", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf("unknown format: %s", text)
	}
	switch s[0] {
	case String:
		return s[1], nil
	case Int:
		return strconv.ParseInt(s[1], 10, 64)
	case Float:
		return strconv.ParseFloat(s[1], 64)
	default:
		return nil, fmt.Errorf("unknown format: %s", text)
	}
}

var (
	String = "string"
	Int    = "int"
	Float  = "float"
)
