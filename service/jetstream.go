package service

import "regexp"

type JetStreamRecord struct {
	bucket     string
	key        string
	value      []byte
	encryption []byte
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

// SetBucket sets the bucket field on a JetStreamRecord
func (j *JetStreamRecord) SetBucket(b string) *JetStreamRecord {
	j.bucket = b
	return j
}

// SetKey sets the key field on a JetStreamRecord
func (j *JetStreamRecord) SetKey(k string) *JetStreamRecord {
	j.key = k
	return j
}

// SetSanitizedKey removes the prefix from the key name on a JetStreamRecord
// This is to keep from having the bucket name duplicated in the subject
func (j *JetStreamRecord) SetSanitizedKey(k string) *JetStreamRecord {
	reg := regexp.MustCompile(`piggybank.secrets.\w+.`)
	subj := reg.ReplaceAllString(k, "${1}")
	j.key = subj
	return j
}

// SetValue sets the value field on a JetStreamRecord
func (j *JetStreamRecord) SetValue(v string) *JetStreamRecord {
	j.value = []byte(v)
	return j
}

// Encrypt encrypts the value of the JetStreamRecord using the encryption key stored in the record
func (j *JetStreamRecord) Encrypt() error {
	v, err := encrypt(j.value, j.encryption)
	if err != nil {
		return err
	}
	j.value = v

	return nil
}

// Decrypt decrypts the value of the JetStreamRecord using the encryption key stored in the record
func (j *JetStreamRecord) Decrypt() error {
	v, err := decrypt(j.value, j.encryption)
	if err != nil {
		return err
	}

	j.value, err = fromBase64(string(v))
	if err != nil {
		return err
	}

	return nil
}

// SetEncryptionKey sets the encryption key in the JetStreamRecord
func (j *JetStreamRecord) SetEncryptionKey(k []byte) *JetStreamRecord {
	j.encryption = k
	return j
}
