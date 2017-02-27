package app

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	String = "string"
	Int    = "int"
	Float  = "float"
)

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
