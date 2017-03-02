package app

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeToAccessor(t *testing.T) {
	type Input struct {
		Format string
		Text   string
	}
	type Expect struct {
		IsNilAccessor bool
		Err           error
	}
	type Test struct {
		Title  string
		Input  Input
		Expect Expect
	}

	table := []Test{
		{
			Title: "empty format",
			Input: Input{
				Format: "",
				Text:   "",
			},
			Expect: Expect{
				true,
				ErrFormatRequired,
			},
		},
		{
			Title: "text success",
			Input: Input{
				Format: "text",
				Text:   "text",
			},
			Expect: Expect{
				false,
				nil,
			},
		},
		{
			Title: "text empty",
			Input: Input{
				Format: "text",
				Text:   "",
			},
			Expect: Expect{
				true,
				ErrEmptyInput,
			},
		},
		{
			Title: "json success",
			Input: Input{
				Format: "json",
				Text:   "[1,2,3]",
			},
			Expect: Expect{
				false,
				nil,
			},
		},
		{
			Title: "json empty",
			Input: Input{
				Format: "json",
				Text:   "",
			},
			Expect: Expect{
				true,
				ErrEmptyInput,
			},
		},
		{
			Title: "yaml success",
			Input: Input{
				Format: "yaml",
				Text:   "aaa: bbb",
			},
			Expect: Expect{
				false,
				nil,
			},
		},
		{
			Title: "yaml empty",
			Input: Input{
				Format: "yaml",
				Text:   "",
			},
			Expect: Expect{
				true,
				ErrEmptyInput,
			},
		},
		{
			Title: "toml success",
			Input: Input{
				Format: "toml",
				Text:   "aaa = 'bbb'",
			},
			Expect: Expect{
				false,
				nil,
			},
		},
		{
			Title: "toml empty",
			Input: Input{
				Format: "toml",
				Text:   "",
			},
			Expect: Expect{
				true,
				ErrEmptyInput,
			},
		},
	}

	for _, test := range table {
		t.Run(test.Title, func(t *testing.T) {
			assert := assert.New(t)

			acc, err := decodeToAccessor(test.Input.Format, bytes.NewBufferString(test.Input.Text))

			if test.Expect.IsNilAccessor {
				assert.Nil(acc)
			} else {
				assert.NotNil(acc)
			}
			assert.Equal(test.Expect.Err, err)
		})
	}
}
