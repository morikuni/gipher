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
	session, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(region),
		},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}
	kms := kms.New(session)
	return awsKMSCryptor{keyID, kms}, nil
}

func (c awsKMSCryptor) Encrypt(text string) (Ciphertext, error) {
	r, err := c.kms.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(c.keyID),
		Plaintext: []byte(text),
	})
	if err != nil {
		return nil, err
	}
	return EncodeCiphertext(r.CiphertextBlob), nil
}

func (c awsKMSCryptor) Decrypt(text Ciphertext) (string, error) {
	ciphertext, err := DecodeCiphertext(text)
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
