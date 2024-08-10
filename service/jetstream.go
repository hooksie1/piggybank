package service

import "regexp"

type JetStreamRecord struct {
	bucket        string
	key           string
	value         []byte
	encryptionKey []byte
}

// NewJSRecord returns a new JetStreamRecord
func NewJSRecord() *JetStreamRecord {
	return &JetStreamRecord{}
}

// Bucket returns the Bucket value for the JetStreamRecord
func (j *JetStreamRecord) Bucket() string {
	return j.bucket
}

// Key returns the Key for the JetStreamRecord
func (j *JetStreamRecord) Key() string {
	return j.key
}

// Value returns the JetStreamRecord value
func (j *JetStreamRecord) Value() []byte {
	return j.value
}

func SanitizeKey(k string) string {
	reg := regexp.MustCompile(`piggybank.secrets.\w+.`)
	return reg.ReplaceAllString(k, "${1}")
}

// Encrypt encrypts the value of the JetStreamRecord using the encryption key stored in the record
func (j *JetStreamRecord) Encrypt() error {
	v, err := encrypt(j.value, j.encryptionKey)
	if err != nil {
		return err
	}
	j.value = v

	return nil
}

// Decrypt decrypts the value of the JetStreamRecord using the encryption key stored in the record
func (j *JetStreamRecord) Decrypt() ([]byte, error) {
	v, err := decrypt(j.value, j.encryptionKey)
	if err != nil {
		return nil, err
	}

	return fromBase64(string(v))
}
