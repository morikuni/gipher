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
	Bool   = "bool"
	Nil    = "nil"
)

func encodeToString(value interface{}) (string, bool) {
	if value == nil {
		return Nil + ":", true
	}

	switch t := value.(type) {
	case string:
		return String + ":" + t, true
	case int:
		return Int + ":" + strconv.FormatInt(int64(t), 10), true
	case int64:
		return Int + ":" + strconv.FormatInt(t, 10), true
	case float32:
		return Float + ":" + strconv.FormatFloat(float64(t), 'E', -1, 64), true
	case float64:
		return Float + ":" + strconv.FormatFloat(t, 'E', -1, 64), true
	case bool:
		return Bool + ":" + strconv.FormatBool(t), true
	default:
		return "", false
	}
}

func decodeFromString(text string) (interface{}, error) {
	s := strings.SplitN(text, ":", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf("invalid format: %s", text)
	}
	switch s[0] {
	case String:
		return s[1], nil
	case Int:
		return strconv.ParseInt(s[1], 10, 64)
	case Float:
		return strconv.ParseFloat(s[1], 64)
	case Bool:
		return strconv.ParseBool(s[1])
	case Nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid format: %s", text)
	}
}
