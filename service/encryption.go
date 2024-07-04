package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
)

// toBase64 takes a byte slice and returns a base64 encoded string of that slice
func toBase64(key []byte) string {
	encoded := base64.RawStdEncoding.EncodeToString(key)

	return encoded
}

// fromBase64 takes a base64 encoded string and returns the decode string
func fromBase64(encoded string) ([]byte, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func generateKey() []byte {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		log.Println(err)
	}

	return key
}

// encrypt takes a plain text secret and a 32 bit key and encrypts
// the secret using the key. It returns the encrypted text or an error.
func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decrypt takes a byte slice and a 32 bit key and decrypts
// the secret using the key. It returns the decrypted value
// or an error.
func decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}
