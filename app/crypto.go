package app

import (
	"errors"
	"fmt"

	"github.com/morikuni/gipher"
)

func createCryptor(cryptor, command, awsRegion, awsKeyID string) (gipher.Cryptor, error) {
	switch cryptor {
	case "":
		return nil, errors.New("cryptor is required")
	case "password":
		return gipher.NewPasswordCryptorWithPrompt()
	case "aws-kms":
		if awsRegion == "" {
			return nil, errors.New("aws-region is required for aws-kms")
		}
		if awsKeyID == "" && command == "encrypt" {
			return nil, errors.New("key-id is required for aws-kms")
		}
		return gipher.NewAWSKMSCryptor(awsRegion, awsKeyID)
	default:
		return nil, fmt.Errorf("unknown cryptor: %q", cryptor)
	}
}

func decrypt(cryptor gipher.Cryptor, value string) (interface{}, error) {
	text, err := cryptor.Decrypt(gipher.Base64String(value))
	if err != nil {
		return "", err
	}
	return decodeFromString(text)
}

func encrypt(cryptor gipher.Cryptor, value interface{}) (string, bool, error) {
	text, shouldSet, err := encodeToString(value)
	if !shouldSet || err != nil {
		return "", shouldSet, err
	}
	cipher, err := cryptor.Encrypt(text)
	return string(cipher), true, err
}
