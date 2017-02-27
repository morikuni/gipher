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
