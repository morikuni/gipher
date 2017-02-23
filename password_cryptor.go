package gipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

type passwordCryptor struct {
	passwordHash []byte
}

func NewPasswordCryptor(password []byte) Cryptor {
	hash := sha256.Sum256(password)
	return passwordCryptor{
		hash[:],
	}
}

func (c passwordCryptor) Encrypt(text string) (Base64String, error) {
	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.passwordHash)
	if err != nil {
		return "", fmt.Errorf("cannot accept the encryption key: %s", err)
	}
	stream := cipher.NewCTR(block, iv)

	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return Base64String(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func (c passwordCryptor) Decrypt(text Base64String) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext as base64: %s", err)
	}
	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)

	iv := ciphertext[:aes.BlockSize]

	block, err := aes.NewCipher(c.passwordHash)
	if err != nil {
		return "", fmt.Errorf("cannot accept the decryption key: %s", err)
	}
	stream := cipher.NewCTR(block, iv)

	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

	return string(plaintext), nil
}
