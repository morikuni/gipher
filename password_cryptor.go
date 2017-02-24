package gipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

var ErrCannotReadPassword = errors.New("cannot read the password. use GIPHER_PASSWORD to set the password if you used stdin for input.")

type passwordCryptor struct {
	passwordHash []byte
}

func NewPasswordCryptor(password []byte) Cryptor {
	hash := sha256.Sum256(password)
	return passwordCryptor{
		hash[:],
	}
}

func NewPasswordCryptorWithPrompt() (Cryptor, error) {
	pass := os.Getenv("GIPHER_PASSWORD")
	if pass != "" {
		return NewPasswordCryptor([]byte(pass)), nil
	}

	fd := int(os.Stdin.Fd())
	if !terminal.IsTerminal(fd) {
		return nil, ErrCannotReadPassword
	}
	fmt.Fprint(os.Stderr, "password:")
	p, err := terminal.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return nil, err
	}
	return NewPasswordCryptor(p), nil
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

	return EncodeBase64(ciphertext), nil
}

func (c passwordCryptor) Decrypt(text Base64String) (string, error) {
	ciphertext, err := DecodeBase64(text)
	if err != nil {
		return "", err
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
