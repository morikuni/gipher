package app

import (
	"fmt"
	"io"
	"regexp"

	"github.com/morikuni/accessor"
	"github.com/spf13/pflag"
)

var DryrunMessage = "THIS FIELD WILL BE CHENGED"

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
	pattern := flag.String("pattern", ".*", `regular expression. only fields matching the pattern are encrypted/decrypted (e.g. "user/items/.*/name").`)
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

	if len(flag.Args()) > 2 {
		fmt.Fprintln(stderr, "too many args")
		return 1
	}

	command := flag.Arg(1)
	if command == "" {
		fmt.Fprintln(stderr, "command is required")
		flag.Usage()
		return 1
	}

	reg, err := regexp.Compile(*pattern)
	if err != nil {
		fmt.Fprintf(stderr, "invalid pattern: %s\n", err)
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

	cryptor, err := createCryptor(*cryptorType, command, *awsRegion, *awsKeyID)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	err = acc.Foreach(func(path accessor.Path, value interface{}) error {
		if !reg.MatchString(path.String()) {
			return nil
		}

		if *dryrun {
			return acc.Set(path, DryrunMessage)
		}

		switch command {
		case "encrypt":
			cipher, shouldSet, err := encrypt(cryptor, value)
			if err != nil {
				return err
			}
			if shouldSet {
				return acc.Set(path, cipher)
			}
		case "decrypt":
			if s, ok := value.(string); ok {
				value, err := decrypt(cryptor, s)
				if err != nil {
					return err
				}
				return acc.Set(path, value)
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
