package app

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeToString(t *testing.T) {
	type Input struct {
		Value interface{}
	}
	type Expect struct {
		String    string
		ShouldSet bool
	}
	type Test struct {
		Title  string
		Input  Input
		Expect Expect
	}

	table := []Test{
		{
			Title: "string",
			Input: Input{"aaa"},
			Expect: Expect{
				String:    "string:aaa",
				ShouldSet: true,
			},
		},
		{
			Title: "int",
			Input: Input{int(math.MaxInt32)},
			Expect: Expect{
				String:    "int:2147483647",
				ShouldSet: true,
			},
		},
		{
			Title: "int64",
			Input: Input{int64(math.MaxInt64)},
			Expect: Expect{
				String:    "int:9223372036854775807",
				ShouldSet: true,
			},
		},
		{
			Title: "float32",
			Input: Input{float32(123123)},
			Expect: Expect{
				String:    "float:1.23123E+05",
				ShouldSet: true,
			},
		},
		{
			Title: "float64",
			Input: Input{float64(123123123123)},
			Expect: Expect{
				String:    "float:1.23123123123E+11",
				ShouldSet: true,
			},
		},
		{
			Title: "nil",
			Input: Input{nil},
			Expect: Expect{
				String:    "nil:",
				ShouldSet: true,
			},
		},
		{
			Title: "other",
			Input: Input{1 + 2i},
			Expect: Expect{
				String:    "",
				ShouldSet: false,
			},
		},
	}

	for _, test := range table {
		t.Run(test.Title, func(t *testing.T) {
			assert := assert.New(t)

			s, b := encodeToString(test.Input.Value)

			assert.Equal(test.Expect.String, s)
			assert.Equal(test.Expect.ShouldSet, b)
		})
	}
}

func TestDecodeFromString(t *testing.T) {
	type Input struct {
		String string
	}
	type Expect struct {
		Value interface{}
		Err   error
	}
	type Test struct {
		Title  string
		Input  Input
		Expect Expect
	}

	table := []Test{
		{
			Title: "string",
			Input: Input{"string:aaa"},
			Expect: Expect{
				Value: "aaa",
				Err:   nil,
			},
		},
		{
			Title: "int",
			Input: Input{"int:2147483647"},
			Expect: Expect{
				Value: int64(2147483647),
				Err:   nil,
			},
		},
		{
			Title: "int64",
			Input: Input{"int:9223372036854775807"},
			Expect: Expect{
				Value: int64(9223372036854775807),
				Err:   nil,
			},
		},
		{
			Title: "float32",
			Input: Input{"float:1.23123E+05"},
			Expect: Expect{
				Value: float64(123123),
				Err:   nil,
			},
		},
		{
			Title: "float64",
			Input: Input{"float:1.23123123123E+11"},
			Expect: Expect{
				Value: float64(123123123123),
				Err:   nil,
			},
		},
		{
			Title: "nil",
			Input: Input{"nil:"},
			Expect: Expect{
				Value: nil,
				Err:   nil,
			},
		},
		{
			Title: "no tag",
			Input: Input{"2147483647"},
			Expect: Expect{
				Value: nil,
				Err:   errors.New("unknown format: 2147483647"),
			},
		},
		{
			Title: "invalid tag",
			Input: Input{"tag:2147483647"},
			Expect: Expect{
				Value: nil,
				Err:   errors.New("unknown format: tag:2147483647"),
			},
		},
	}

	for _, test := range table {
		t.Run(test.Title, func(t *testing.T) {
			assert := assert.New(t)

			v, err := decodeFromString(test.Input.String)

			assert.Equal(test.Expect.Value, v)
			assert.Equal(test.Expect.Err, err)
		})
	}
}
