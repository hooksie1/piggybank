package server

import (
	"bytes"
	"fmt"

	"go.etcd.io/bbolt"
)

// BoltRecord is a record that holds the values
// for a BoltDB Bucket, Key, and VAlue.
type BoltRecord struct {
	Bucket string
	Key    []byte
	Value  []byte
}

type BoltOption func(*BoltRecord) (*BoltRecord, error)

// NewBoltRecord takes optional values and either returns an empty
// BoltRecord or a record containing the optional values.
func NewBoltRecord(opts ...BoltOption) (*BoltRecord, error) {
	b := &BoltRecord{}
	var err error

	for _, opt := range opts {
		b, err = opt(b)
		if err != nil {
			return nil, fmt.Errorf("Error creating Bolt Record: %s", err)
		}
	}

	return b, nil
}

// BoltBucket sets the bucket as the string parameter
func BoltBucket(bucket string) BoltOption {
	return func(b *BoltRecord) (*BoltRecord, error) {
		b.Bucket = bucket
		return b, nil
	}
}

// BoltKey sets the key as the byte slice parameter
func BoltKey(key []byte) BoltOption {
	return func(b *BoltRecord) (*BoltRecord, error) {
		b.Key = key
		return b, nil
	}
}

// BoltValue sets the value as the byte slice parameter
func BoltValue(value []byte) BoltOption {
	return func(b *BoltRecord) (*BoltRecord, error) {
		b.Value = value
		return b, nil
	}
}

// AddRecord creates a bucket and a key value pair inside of the bucket
// using a record for values.
func (r *BoltRecord) AddRecord() error {
	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		return fmt.Errorf("error opening database, %s", err)
	}

	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(r.Bucket))
		if err != nil {
			return fmt.Errorf("error creating bucket %s", err)
		}

		err = b.Put(r.Key, r.Value)
		if err != nil {
			return fmt.Errorf("error adding key/value %s", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetPlainRecord specifically looks for an unencrypted key name
func (r *BoltRecord) GetPlainRecord(db *bbolt.DB) error {

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(r.Bucket))
		v := b.Get(r.Key)
		r.Value = v

		return nil
	})
	if err != nil {
		return err
	}

	return nil

}

// GetEncryptedRecord reads only encrypted keys in a bucket.
func (r *BoltRecord) GetEncryptedRecord(db *bbolt.DB) error {

	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(r.Bucket))
		if b == nil {
			return fmt.Errorf("record not found")
		}
		// loop through the bucket's k/v pairs and decrypt the username
		// to find the correct value.
		b.ForEach(func(k, v []byte) error {
			plainUser, err := decrypt(k, masterPass)
			if err != nil {
				return err
			}
			if bytes.Equal(plainUser, r.Key) {
				r.Key = k
				r.Value = v
			}
			return nil
		})

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// GetRecord calls either the GetPlainRecord or GetEncryptedRecord
// depending on whether the bucket name.
func (r *BoltRecord) GetRecord() error {

	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		return fmt.Errorf("error opening database: %s", err)
	}

	defer db.Close()

	if r.Bucket == "Users" {
		if err := r.GetPlainRecord(db); err != nil {
			return err
		}
		return nil
	}

	if r.Bucket == "initialized" {
		if err := r.GetPlainRecord(db); err != nil {
			return err
		}

		return nil
	}

	if err := r.GetEncryptedRecord(db); err != nil {
		return err
	}

	return nil

}

// DeleteRecord calls either DeletetPlainRecord or
// DeleteEncryptedRecord depending on the name of the bucket.
func (r *BoltRecord) DeleteRecord() error {
	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		return fmt.Errorf("error opening database: %s", err)
	}

	defer db.Close()

	if r.Bucket == "Users" {
		if err := r.DeletePlainRecord(db); err != nil {
			return err
		}

		return nil
	}

	if err := r.DeleteEncryptedRecord(db); err != nil {
		return err
	}

	return nil

}

// DeletePlainRecord deletes keys with only plaintext values. These
// are things like user records.
func (r *BoltRecord) DeletePlainRecord(db *bbolt.DB) error {
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(r.Bucket))
		b.Delete(r.Key)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteEncryptedRecord only deletes keys that are encrypted. These
// are things like credentials.
func (r *BoltRecord) DeleteEncryptedRecord(db *bbolt.DB) error {

	if err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(r.Bucket))
		if b == nil {
			return fmt.Errorf("record not found")
		}
		// loop through the bucket's k/v pairs and decrypt the username
		// to find the correct value.
		b.ForEach(func(k, v []byte) error {
			plainUser, err := decrypt(k, masterPass)
			if err != nil {
				return err
			}
			if bytes.Equal(plainUser, r.Key) {
				b.Delete(k)
			}
			return nil
		})
		return nil
	}); err != nil {
		return err
	}

	return nil
}
