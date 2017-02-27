package gipher

// Cryptor encrypts/decrypts a text.
type Cryptor interface {
	// Encrypt encrypts a text and encodes it by base64.
	Encrypt(plaintext string) (Ciphertext, error)

	// Decrypt decodes a text by base64 and decrypts it.
	Decrypt(ciphertext Ciphertext) (string, error)
}
