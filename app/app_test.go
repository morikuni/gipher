package app

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/morikuni/gipher"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	type Input struct {
		Args  string
		Stdin string
		Env   map[string]string
	}
	type Expect struct {
		ExitCode int
		Stdout   string
		Stderr   string
	}
	type Test struct {
		Title  string
		Input  Input
		Expect Expect
	}

	passwordEnv := map[string]string{
		"GIPHER_PASSWORD": "aaaa",
	}

	table := []Test{
		{
			Title: "help message",
			Input: Input{
				Args:  "gipher -h",
				Stdin: "",
			},
			Expect: Expect{
				ExitCode: 0,
				Stdout:   `\A\z`,
				Stderr:   `Usage: gipher`,
			},
		},
		{
			Title: "too many args",
			Input: Input{
				Args:  "gipher encrypt hoge",
				Stdin: "",
			},
			Expect: Expect{
				ExitCode: 1,
				Stdout:   `\A\z`,
				Stderr:   `too many args`,
			},
		},
		{
			Title: "empty args",
			Input: Input{
				Args:  "gipher",
				Stdin: "",
			},
			Expect: Expect{
				ExitCode: 1,
				Stdout:   `\A\z`,
				Stderr:   `command is required`,
			},
		},
		{
			Title: "invalid pattern",
			Input: Input{
				Args:  "gipher encrypt --pattern (",
				Stdin: "aaa",
				Env:   passwordEnv,
			},
			Expect: Expect{
				ExitCode: 1,
				Stdout:   `\A\z`,
				Stderr:   `invalid pattern: `,
			},
		},
		{
			Title: "no such file",
			Input: Input{
				Args:  "gipher encrypt -f no_such_file.txt",
				Stdin: "",
			},
			Expect: Expect{
				ExitCode: 1,
				Stdout:   `\A\z`,
				Stderr:   `open no_such_file.txt: no such file or directory`,
			},
		},
		{
			Title: "empty input",
			Input: Input{
				Args:  "gipher encrypt",
				Stdin: "",
			},
			Expect: Expect{
				ExitCode: 1,
				Stdout:   `\A\z`,
				Stderr:   `input is empty`,
			},
		},
		{
			Title: "unknown command",
			Input: Input{
				Args:  "gipher test",
				Stdin: "aaa",
				Env:   passwordEnv,
			},
			Expect: Expect{
				ExitCode: 1,
				Stdout:   `\A\z`,
				Stderr:   `unknown command: test`,
			},
		},
		{
			Title: "dryrun",
			Input: Input{
				Args: "gipher encrypt --format json --pattern name --dryrun",
				Stdin: `{
						"name": "Alice",
						"age": 18
					}
				`,
				Env: passwordEnv,
			},
			Expect: Expect{
				ExitCode: 0,
				Stdout:   fmt.Sprintf(`{"age":18,"name":"%s"}`, DryrunMessage),
				Stderr:   `\A\z`,
			},
		},
		{
			Title: "encrypt: success password",
			Input: Input{
				Args: "gipher encrypt --format json --pattern name",
				Stdin: `{
						"name": "Alice",
						"age": 18
					}
				`,
				Env: passwordEnv,
			},
			Expect: Expect{
				ExitCode: 0,
				Stdout:   `{"age":18,"name":"[0-9a-zA-Z+=/]{40}"}`,
				Stderr:   `\A\z`,
			},
		},
		{
			Title: "decrypt: success password",
			Input: Input{
				Args: "gipher decrypt --format json --pattern name",
				Stdin: `{
						"name": "R1lyLATIeGJC5UYEGne+KtOr4VzWsn0qeqxjJw==",
						"age": 18
					}
				`,
				Env: passwordEnv,
			},
			Expect: Expect{
				ExitCode: 0,
				Stdout:   `{"age":18,"name":"Alice"}`,
				Stderr:   `\A\z`,
			},
		},
	}

	if profile := os.Getenv("TEST_AWS_PROFILE"); profile != "" {
		keyID := os.Getenv("TEST_AWS_KEY_ID")
		region := os.Getenv("TEST_AWS_REGION")

		os.Setenv("AWS_PROFILE", profile)
		kms, err := gipher.NewAWSKMSCryptor(region, keyID)
		if err != nil {
			t.Error(err)
			return
		}
		os.Unsetenv("AWS_PROFILE")

		cipher, err := kms.Encrypt("string:Alice")
		if err != nil {
			t.Error(err)
			return
		}

		table = append(table, Test{
			Title: "encrypt: success aws kms",
			Input: Input{
				Args: "gipher encrypt --format json --pattern name --cryptor aws-kms --aws-key-id " + keyID + " --aws-region " + region,
				Stdin: `{
						"name": "Alice",
						"age": 18
					}
				`,
				Env: map[string]string{
					"AWS_PROFILE": profile,
				},
			},
			Expect: Expect{
				ExitCode: 0,
				Stdout:   `{"age":18,"name":"[0-9a-zA-Z+=/]{100,}"}`,
				Stderr:   `\A\z`,
			},
		})
		table = append(table, Test{
			Title: "decrypt: success aws kms",
			Input: Input{
				Args: "gipher decrypt --format json --pattern name --cryptor aws-kms --aws-region " + region,
				Stdin: fmt.Sprintf(`{
						"name": "%s",
						"age": 18
					}
				`, string(cipher)),
				Env: map[string]string{
					"AWS_PROFILE": profile,
				},
			},
			Expect: Expect{
				ExitCode: 0,
				Stdout:   `{"age":18,"name":"Alice"}`,
				Stderr:   `\A\z`,
			},
		})
	}

	for _, test := range table {
		t.Run(test.Title, func(t *testing.T) {
			assert := assert.New(t)

			for k, v := range test.Input.Env {
				os.Setenv(k, v)
			}

			t.Log(test.Input.Args)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			exitCode := NewApp().Run(strings.Fields(test.Input.Args), strings.NewReader(test.Input.Stdin), stdout, stderr)

			assert.Equal(test.Expect.ExitCode, exitCode)
			assert.Regexp(test.Expect.Stdout, stdout.String())
			assert.Regexp(test.Expect.Stderr, stderr.String())

			for k := range test.Input.Env {
				os.Unsetenv(k)
			}
		})
	}
}
