package gipher

import (
	"encoding/base64"
	"fmt"
)

// Base64String is base64-encoded string
type Base64String string

func EncodeBase64(bs []byte) Base64String {
	return Base64String(base64.StdEncoding.EncodeToString(bs))
}

func DecodeBase64(text Base64String) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext as base64: %s", err)
	}
	return ciphertext, nil
}
