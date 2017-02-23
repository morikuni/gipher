package gipher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordCryptor(t *testing.T) {
	type Input struct {
		Plaintext string
		Password  string
	}
	type Test struct {
		Title string
		Input Input
	}

	table := []Test{
		{
			Title: "success",
			Input: Input{
				Plaintext: "hello world",
				Password:  "nice password",
			},
		},
		{
			Title: "success even if password is empty",
			Input: Input{
				Plaintext: "gipher",
				Password:  "",
			},
		},
	}

	for _, test := range table {
		t.Run(test.Title, func(t *testing.T) {
			assert := assert.New(t)

			encrypter := NewPasswordCryptor([]byte(test.Input.Password))
			cipher, err := encrypter.Encrypt(test.Input.Plaintext)
			assert.Nil(err)

			decryptor := NewPasswordCryptor([]byte(test.Input.Password))
			text, err := decryptor.Decrypt(cipher)
			assert.Nil(err)

			assert.Equal(test.Input.Plaintext, text)
		})
	}
}
