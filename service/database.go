package service

import (
	"crypto/aes"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	databaseInitSubject   = "piggybank.database.initialize"
	databaseUnlockSubject = "piggybank.database.unlock"
	databaseLockSubject   = "piggybank.database.lock"
	databaseStatusSubject = "piggybank.database.status"
)

type KV interface {
	Bucket() string
	Key() string
	Value() []byte
	Encrypt() error
}

type Watcher interface {
	Watch()
}

type Backend interface {
	Watcher
}

// initialize sets the initialization key. Once this is set it does not need to be run again, unless you lose the encryption key.
// If you lose the encryption key, everything is lost.
func (a *AppContext) initialize() ([]byte, error) {
	kv := NewJSRecord().SetBucket(piggyBucket).SetKey("init")

	_, err := a.GetRecord(kv)
	if err != nil && err != nats.ErrKeyNotFound {
		return nil, err
	}

	if err == nil {
		return nil, NewClientError(fmt.Errorf("database already initialized"), 400)
	}

	key, random := generateKey(), generatePass()

	record := NewJSRecord().SetEncryptionKey(key).SetBucket(piggyBucket).SetKey("init").SetValue(random)

	if err := record.Encrypt(); err != nil {
		return nil, err
	}

	if err := a.AddRecord(record); err != nil {
		return nil, err
	}

	return key, nil

}

func (a *AppContext) Unlock(k KV) error {
	key, err := fromBase64(string(k.Value()))
	if err != nil {
		return err
	}

	val, err := a.GetRecord(k)
	if err != nil {
		return err
	}

	_, err = decrypt(val, key)
	if err != nil {
		return err
	}

	databaseKey = key

	return nil
}

func (a *AppContext) unlock(data []byte) error {
	var key DatabaseKey

	if databaseKey != nil {
		return NewClientError(fmt.Errorf("database already unlocked"), 400)
	}

	if err := json.Unmarshal(data, &key); err != nil {
		return err
	}

	if key.DBKey == "" || len(key.DBKey) < aes.BlockSize {
		return NewClientError(fmt.Errorf("key is too short"), 400)
	}

	kv := NewJSRecord().SetBucket(piggyBucket).SetKey("init").SetValue(key.DBKey)

	if err := a.Unlock(kv); err != nil {
		return fmt.Errorf("error unlocking database: %v", err)
	}

	return nil
}

// addRecord wraps AddRecord by encrypting the data first and handling responses
func (a *AppContext) addRecord(k KV) error {

	if err := k.Encrypt(); err != nil {
		return err
	}

	if err := a.AddRecord(k); err != nil {
		return err
	}

	return nil
}

func (a *AppContext) AddRecord(k KV) error {
	_, err := a.KV.Put(k.Key(), k.Value())
	if err != nil {
		return err
	}

	return nil
}

// getRecord wraps GetRecord by decrypting the returned value and handling resposnes.
func (a *AppContext) getRecord(k KV) ([]byte, error) {
	data, err := a.GetRecord(k)
	if err != nil && err == nats.ErrKeyNotFound {
		return nil, NewClientError(fmt.Errorf("key not found"), 404)
	}

	if err != nil && err != nats.ErrKeyNotFound {
		return nil, err
	}

	decrypted, err := decrypt(data, databaseKey)
	if err != nil {
		return nil, err
	}

	return decrypted, nil

}

func (a *AppContext) GetRecord(k KV) ([]byte, error) {
	v, err := a.KV.Get(k.Key())
	if err != nil {
		return nil, err
	}

	return v.Value(), nil
}

// deleteRecord wraps DeleteRecord and handles responses.
func (a *AppContext) deleteRecord(k KV) error {
	err := a.DeleteRecord(k)
	if err != nil && err == nats.ErrKeyNotFound {
		return NewClientError(fmt.Errorf("key not found"), 400)
	}

	if err != nil && err != nats.ErrKeyNotFound {
		return err
	}

	return nil
}

func (a *AppContext) DeleteRecord(k KV) error {
	return a.KV.Delete(k.Key())
}
