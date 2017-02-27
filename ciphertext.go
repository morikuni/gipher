package gipher

import (
	"encoding/base64"
	"fmt"
)

// Ciphertext is base64-encoded string
type Ciphertext []byte

func EncodeCiphertext(bs []byte) Ciphertext {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(bs)))
	base64.StdEncoding.Encode(buf, bs)
	return Ciphertext(buf)
}

func DecodeCiphertext(text Ciphertext) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext as base64: %s", err)
	}
	return ciphertext, nil
}
