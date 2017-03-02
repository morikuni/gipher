package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/morikuni/accessor"
	yaml "gopkg.in/yaml.v2"
)

var (
	ErrEmptyInput     = errors.New("input is empty")
	ErrFormatRequired = errors.New("format is required")
)

func ErrUnknownFormat(format string) error {
	return fmt.Errorf("unknown format: %q", format)
}

func decodeToAccessor(format string, input io.Reader) (accessor.Accessor, error) {
	switch format {
	case "":
		return nil, ErrFormatRequired
	case "json":
		var obj interface{}
		err := json.NewDecoder(input).Decode(&obj)
		if err != nil {
			if err == io.EOF {
				return nil, ErrEmptyInput
			}
			return nil, err
		}
		return accessor.NewAccessor(obj)
	case "yaml":
		bs, err := ioutil.ReadAll(input)
		if err != nil {
			return nil, err
		}
		if len(bs) == 0 {
			return nil, ErrEmptyInput
		}
		var obj interface{}
		err = yaml.Unmarshal(bs, &obj)
		if err != nil {
			return nil, err
		}
		return accessor.NewAccessor(obj)
	case "toml":
		bs, err := ioutil.ReadAll(input)
		if err != nil {
			return nil, err
		}
		if len(bs) == 0 {
			return nil, ErrEmptyInput
		}
		var obj interface{}
		_, err = toml.Decode(string(bs), &obj)
		if err != nil {
			return nil, err
		}
		return accessor.NewAccessor(obj)
	case "text":
		bs, err := ioutil.ReadAll(input)
		if err != nil {
			return nil, err
		}
		if len(bs) == 0 {
			return nil, ErrEmptyInput
		}
		return accessor.NewAccessor(string(bs))
	default:
		return nil, ErrUnknownFormat(format)
	}
}

func encodeAccessor(format string, output io.Writer, acc accessor.Accessor) error {
	switch format {
	case "":
		return ErrFormatRequired
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
