package service

import (
	"bytes"
	"fmt"

	"github.com/nats-io/nats.go"
)

type rotatedKV struct {
	subject string
	value   []byte
	rotated bool
	oldKey  []byte
	newKey  []byte
}

func buildKeys(newKey []byte, w nats.KeyWatcher) []rotatedKV {
	kvs := []rotatedKV{}
	for v := range w.Updates() {
		if v == nil {
			break
		}

		kvs = append(kvs, rotatedKV{
			subject: v.Key(),
			value:   v.Value(),
			oldKey:  databaseKey,
			newKey:  newKey,
		})
	}

	return kvs
}

func (a *AppContext) Rotate(currentKey string) ([]byte, error) {
	currentKeyBytes, err := fromBase64(currentKey)
	if err != nil {
		return nil, NewClientError(fmt.Errorf("%v", err), 400)
	}

	if !bytes.Equal(currentKeyBytes, databaseKey) {
		return nil, NewClientError(fmt.Errorf("current database key does not match"), 401)
	}

	a.logger.Info("generating new key")
	newKey := generateKey()

	w, err := a.KV.WatchAll(nats.IgnoreDeletes())
	if err != nil {
		return nil, err
	}

	kvs := buildKeys(newKey, w)

	updated, err := a.rotateKey(kvs)
	if err != nil {
		errs := a.rollbackKey(updated)
		fmt.Println(errs)
		return nil, nil
	}

	databaseKey = newKey

	return []byte(newKey), nil
}

func (a *AppContext) rollbackKey(kvs []rotatedKV) []error {
	logger := a.logger.WithContext(map[string]string{"rotation_step": "rollback"})
	var errs []error
	for _, v := range kvs {
		if v.rotated == true {
			logger.Infof("rolling back secret: %s", v.subject)
			record := NewJSRecord().SetEncryptionKey(v.oldKey).SetBucket(piggyBucket).SetKey(v.subject)

			data, err := a.getRecord(record, v.newKey)
			if err != nil {
				logger.Errorf("error in getting secret %s: %v", v.subject, err)
				continue
			}

			record.SetValue(string(data))

			if err := a.addRecord(record); err != nil {
				errs = append(errs, err)
				logger.Errorf("error rolling back encryption key on secret %s: %v", v.subject, err)
				continue
			}
		}
	}

	return errs
}

func (a *AppContext) rotateKey(kvs []rotatedKV) ([]rotatedKV, error) {
	logger := a.logger.WithContext(map[string]string{"rotation_step": "rotate"})
	for k, v := range kvs {
		logger.Infof("re-encrypting secret %s", v.subject)

		record := NewJSRecord().SetEncryptionKey(v.newKey).SetBucket(piggyBucket).SetKey(v.subject).SetValue(string(v.value))

		data, err := a.getRecord(record, v.oldKey)
		if err != nil {
			logger.Errorf("key rotation error in getting secret %s: %v", v.subject, err)
			return kvs, err
		}

		record.SetValue(string(data))

		if err := a.addRecord(record); err != nil {
			logger.Errorf("key rotation error updating secret %s: %v", v.subject, err)
			return kvs, err
		}
		kvs[k].rotated = true
	}

	return kvs, nil

}
