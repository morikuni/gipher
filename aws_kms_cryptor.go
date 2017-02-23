package gipher

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

type awsKMSCryptor struct {
	keyID string
	kms   *kms.KMS
}

func NewAWSKMSCryptor(region string, keyID string) (Cryptor, error) {
	session, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	kms := kms.New(session)
	return awsKMSCryptor{keyID, kms}, nil
}

func (c awsKMSCryptor) Encrypt(text string) (Base64String, error) {
	r, err := c.kms.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(c.keyID),
		Plaintext: []byte(text),
	})
	if err != nil {
		return "", err
	}
	return EncodeBase64(r.CiphertextBlob), nil
}

func (c awsKMSCryptor) Decrypt(text Base64String) (string, error) {
	ciphertext, err := DecodeBase64(text)
	if err != nil {
		return "", err
	}

	r, err := c.kms.Decrypt(&kms.DecryptInput{
		CiphertextBlob: ciphertext,
	})
	if err != nil {
		return "", err
	}
	return string(r.Plaintext), nil
}
