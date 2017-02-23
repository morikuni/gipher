package gipher

// Cryptor encrypts/decrypts a text.
type Cryptor interface {
	// Encrypt encrypts a text and encodes it by base64.
	Encrypt(plaintext string) (Base64String, error)

	// Decrypt decodes a text by base64 and decrypts it.
	Decrypt(ciphertext Base64String) (string, error)
}
