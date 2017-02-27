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

var (
	ErrCannotReadPassword = errors.New("cannot read the password. use GIPHER_PASSWORD to set the password if you did not use a terminal.")
	ErrPasswordIsEmpty    = errors.New("password is empty")
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

func NewPasswordCryptorWithPrompt() (Cryptor, error) {
	pass := os.Getenv("GIPHER_PASSWORD")
	if pass != "" {
		return NewPasswordCryptor([]byte(pass)), nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		return nil, ErrCannotReadPassword
	}
	defer tty.Close()

	fmt.Fprint(tty, "password:")
	p, err := terminal.ReadPassword(int(tty.Fd()))
	fmt.Fprintln(tty)
	if err != nil {
		return nil, err
	}
	if len(p) == 0 {
		return nil, ErrPasswordIsEmpty
	}
	return NewPasswordCryptor(p), nil
}

func (c passwordCryptor) Encrypt(text string) (Ciphertext, error) {
	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(c.passwordHash)
	if err != nil {
		return nil, fmt.Errorf("cannot accept the encryption key: %s", err)
	}
	stream := cipher.NewCTR(block, iv)

	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return EncodeCiphertext(ciphertext), nil
}

func (c passwordCryptor) Decrypt(text Ciphertext) (string, error) {
	ciphertext, err := DecodeCiphertext(text)
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
