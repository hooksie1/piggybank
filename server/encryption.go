package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func toBase64(key []byte) string {
	encoded := base64.RawStdEncoding.EncodeToString(key)

	return encoded
}

func fromBase64(encoded string) []byte {
	decoded, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		log.Println(err)
	}

	return decoded
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

// checkLength checks the length of the master password and
// returns an error if it is too short.
func checkLength() error {
	if masterPass == nil {
		return fmt.Errorf("master password is empty")
	}

	if len(masterPass) < aes.BlockSize {
		return fmt.Errorf("master password is too short")
	}

	return nil
}

// checkMasterPass takes a byte slice and tries to decrypt
// the initialized value in the database with it.
func checkMasterPass(pass []byte) error {

	r, err := NewBoltRecord(
		BoltBucket("initialized"),
		BoltKey([]byte("init")),
	)
	if err != nil {
		return err
	}

	if err := ReadRecord(r); err != nil {
		return err
	}

	_, err = decrypt(r.Value, pass)
	if err != nil {
		return err
	}

	return nil
}

func unlockSystem(w http.ResponseWriter, r *http.Request) error {
	var password MasterPass

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading data: %s", err)
	}

	if err = json.Unmarshal(body, &password); err != nil {
		return fmt.Errorf("error in json data: %s", err)
	}

	if password.MasterPassword == "" || len(password.MasterPassword) < aes.BlockSize {
		return NewHTTPError(err, http.StatusUnauthorized, "User or password incorrect")

	}

	decoded := fromBase64(password.MasterPassword)

	if err := checkMasterPass(decoded); err != nil {
		return NewHTTPError(err, http.StatusUnauthorized, "User or password incorrect")
	}

	masterPass = decoded

	return nil
}
