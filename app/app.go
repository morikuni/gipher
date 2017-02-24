package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

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
		switch command {
		case "encrypt":
			if s, ok := value.(string); ok {
				base64, err := cryptor.Encrypt(s)
				if err != nil {
					return err
				}
				return target.Set(path, string(base64))
			}
		case "decrypt":
			if s, ok := value.(string); ok {
				text, err := cryptor.Decrypt(gipher.Base64String(s))
				if err != nil {
					return err
				}
				return target.Set(path, text)
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
		return nil, err
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
