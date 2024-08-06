package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

var (
	databaseKey []byte
	piggyBucket = "piggybank"
)

type AppContext struct {
	KV     nats.KeyValue
	logger *logr.Logger
}

// ResponseMessage holds a response to the caller
type ResponseMessage struct {
	Details string `json:"details,omitempty"`
}

type RotateRequest struct {
	CurrentKey string `json:"current_key"`
}

// SecretHandler wraps any secret handlers to check if database is currently locked
func SecretHandler(a AppHandlerFunc) AppHandlerFunc {
	return func(r micro.Request, app AppContext) error {
		if databaseKey == nil {
			return NewClientError(fmt.Errorf("database locked"), 403)
		}
		return a(r, app)
	}

}

func Lock(r micro.Request, app AppContext) error {
	databaseKey = nil
	return r.RespondJSON(ResponseMessage{Details: "database locked"})
}

func Initialize(r micro.Request, app AppContext) error {
	app.logger.Info("initializing database")
	data, err := app.initialize()
	if err != nil {
		return err

	}
	return r.RespondJSON(ResponseMessage{Details: toBase64(data)})
}

func RotateKey(r micro.Request, app AppContext) error {
	var rotateReq RotateRequest

	if err := json.Unmarshal(r.Data(), &rotateReq); err != nil {
		return NewClientError(fmt.Errorf("bad request"), 400)
	}

	if rotateReq.CurrentKey == "" {
		return NewClientError(fmt.Errorf("current db key required"), 400)
	}

	app.logger.Info("rotating encryption key")
	data, err := app.Rotate(rotateReq.CurrentKey)
	if err != nil {
		return err
	}

	return r.RespondJSON(ResponseMessage{Details: toBase64(data)})
}

func Unlock(r micro.Request, app AppContext) error {
	var unlocked bool
	kv := NewJSRecord().SetBucket(piggyBucket).SetKey("init")
	if databaseKey != nil {
		unlocked = true
	}

	if unlocked {
		return NewClientError(fmt.Errorf("database already unlocked"), 400)
	}

	_, err := app.GetRecord(kv)
	if err != nil && err != nats.ErrKeyNotFound {
		return err
	}

	if err == nats.ErrKeyNotFound {
		return NewClientError(fmt.Errorf("database not initialized"), 400)
	}

	app.logger.Info("unlocking database")
	if err := app.unlock(r.Data()); err != nil {
		return err
	}

	return r.RespondJSON(ResponseMessage{Details: "database successfully unlocked"})
}

// Wrap Status in secret handler so it will catch locked requests
func Status(r micro.Request, app AppContext) error {
	return r.RespondJSON(ResponseMessage{Details: "database unlocked"})
}

func GetRecord(r micro.Request, app AppContext) error {
	record := NewJSRecord().SetBucket(piggyBucket).SetSanitizedKey(r.Subject())
	decrypted, err := app.getRecord(record)
	if err != nil {
		return err
	}

	return r.RespondJSON(ResponseMessage{Details: string(decrypted)})
}

func AddRecord(r micro.Request, app AppContext) error {
	record := NewJSRecord().SetBucket(piggyBucket).SetSanitizedKey(r.Subject())
	record.SetValue(string(r.Data()))
	record.SetEncryptionKey(databaseKey)
	if err := app.addRecord(record); err != nil {
		return err
	}

	return r.RespondJSON(ResponseMessage{Details: "successfully stored secret"})
}

func DeleteRecord(r micro.Request, app AppContext) error {
	record := NewJSRecord().SetBucket(piggyBucket).SetSanitizedKey(r.Subject())
	if err := app.deleteRecord(record); err != nil {
		return err
	}

	return r.RespondJSON(ResponseMessage{Details: "successfully deleted secret"})

}

func WatchForConfig(logger *logr.Logger, js nats.JetStreamContext) {
	kv, err := js.KeyValue("configs")
	if err != nil {
		logger.Fatal(err)
	}

	w, err := kv.Watch("piggybank.log_level")
	if err != nil {
		logger.Fatal(err)
	}

	for val := range w.Updates() {
		if val == nil {
			continue
		}

		level := string(val.Value())
		if level == "info" {
			logger.Level = logr.InfoLevel
		}

		if level == "error" {
			logger.Level = logr.ErrorLevel
		}

		if level == "debug" {
			logger.Level = logr.DebugLevel
		}

		logger.Infof("set log level to %s", level)
	}

	time.Sleep(5 * time.Second)
}
