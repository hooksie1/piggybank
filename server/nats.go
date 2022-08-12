package server

import (
	"crypto/aes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/nats-io/nats.go"
)

var (
	ErrInitialized    = errors.New("database already initialized")
	InitializeSubject = fmt.Sprintf("%s.database.initialize", piggyBucket)
	UnlockSubject     = fmt.Sprintf("%s.database.unlock", piggyBucket)
	LockSubject       = fmt.Sprintf("%s.database.lock", piggyBucket)
)

// ResponseMessage holds a response to the caller
type ResponseMessage struct {
	Details string `json:"details,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    int
}

// NatsBackend holds the information for the NATS connection. Fullfills the backend interface.
type NatsBackend struct {
	Servers string
	Options []nats.Option
	Conn    *nats.Conn
	JS      nats.JetStreamContext
}

// NewNatsBackend returns a NATS backend with the supplied connection string and options
func NewNatsBackend(s string, opts []nats.Option) *NatsBackend {
	return &NatsBackend{
		Servers: s,
		Options: opts,
	}
}

// Connect connects to the NATS servers
func (n *NatsBackend) Connect() error {
	nc, err := nats.Connect(n.Servers, n.Options...)
	if err != nil {
		return err
	}

	n.Conn = nc
	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	n.JS = js

	return nil
}

// Watch is the entrypoint for new messages
func (n *NatsBackend) Watch() {
	subject := "piggybank.>"
	log.Printf("watching for requests on %s", subject)
	_, err := n.Conn.Subscribe(subject, n.HandleAndLogRequests)
	if err != nil {
		log.Printf("Error in piggybank service: %s", err)
	}
}

// HandleAndLogRequests just logs the request as it comes in
func (n *NatsBackend) HandleAndLogRequests(m *nats.Msg) {
	log.Printf("%s request sent on subject %s", m.Header.Get("method"), m.Subject)

	n.HandleRequests(m)
}

// HandleRequests determins if the request is a database action (initialization or locking/unlocking) based on the subject name.
func (n *NatsBackend) HandleRequests(m *nats.Msg) {

	if m.Subject == InitializeSubject || m.Subject == UnlockSubject || m.Subject == LockSubject {
		msg := n.HandleDatabaseAction(m)
		if err := m.RespondMsg(msg.Marshal()); err != nil {
			log.Printf("error responding to message: %s", err)
		}
		return
	}

	if databaseKey == nil {
		msg := ResponseMessage{
			Code:  403,
			Error: "database locked",
		}
		if err := m.RespondMsg(msg.Marshal()); err != nil {
			log.Printf("error responding to message: %s", err)
		}
		return
	}

	msg := n.HandleKeyAction(m)
	if err := m.RespondMsg(msg.Marshal()); err != nil {
		log.Printf("error responding to message: %s", err)
	}
}

// HandleKeyAction handles the action for the requested key based on the method in the header
func (n *NatsBackend) HandleKeyAction(m *nats.Msg) *ResponseMessage {
	var msg *ResponseMessage
	record := NewJSRecord().SetBucket(piggyBucket).SetSanitizedKey(m.Subject)

	switch m.Header.Get("method") {
	case "post":
		record.SetValue(string(m.Data))
		record.SetEncryptionKey(databaseKey)
		msg = n.addRecord(record)
	case "get":
		msg = n.getRecord(record)
	case "delete":
		msg = n.deleteRecord(record)
	default:
		msg = n.getRecord(record)
	}

	return msg

}

func (n *NatsBackend) HandleDatabaseAction(m *nats.Msg) *ResponseMessage {
	var unlocked bool
	kv := NewJSRecord().SetBucket(piggyBucket).SetKey("init")
	if databaseKey != nil {
		unlocked = true
	}

	if m.Subject == LockSubject {
		n.Lock()
		return &ResponseMessage{
			Code:    200,
			Details: "database locked",
		}
	}

	if m.Subject == UnlockSubject && unlocked {
		return &ResponseMessage{
			Code:  400,
			Error: "database already unlocked",
		}
	}

	_, err := n.GetRecord(kv)
	if err != nil && err != nats.ErrKeyNotFound {
		log.Println(err)
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	if m.Subject == UnlockSubject && err == nats.ErrKeyNotFound {
		return &ResponseMessage{
			Code:  400,
			Error: "database not initialized",
		}
	}

	if m.Subject == UnlockSubject {
		return n.unlock(m)
	}

	return n.initialize()

}

// initialize sets the initialization key. Once this is set it does not need to be run again, unless you lose the encryption key.
// If you lose the encryption key, everything is lost.
func (n *NatsBackend) initialize() *ResponseMessage {
	kv := NewJSRecord().SetBucket(piggyBucket).SetKey("init")

	_, err := n.GetRecord(kv)
	if err != nil && err != nats.ErrKeyNotFound {
		log.Println(err)
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	if err == nil {
		return &ResponseMessage{
			Code:  400,
			Error: "database already initialized",
		}
	}

	data, err := n.Initialize()
	if err != nil {
		log.Println(err)
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	b64 := toBase64(data)

	return &ResponseMessage{
		Details: b64,
		Code:    200,
	}
}

// Initialize generates a password and encrypts it with the generated key. The key is returned to the caller. If the key is lost,
// the data cannot be recovered.
func (n *NatsBackend) Initialize() ([]byte, error) {
	key, random := generateKey(), generatePass()

	record := NewJSRecord().SetEncryptionKey(key).SetBucket(piggyBucket).SetKey("init").SetValue(random)

	if err := record.Encrypt(); err != nil {
		return nil, err
	}

	if err := n.AddRecord(record); err != nil {
		return nil, err
	}

	return key, nil

}

// addRecord wraps AddRecord by encrypting the data first and handling responses
func (n *NatsBackend) addRecord(k KV) *ResponseMessage {

	if err := k.Encrypt(); err != nil {
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	if err := n.AddRecord(k); err != nil {
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	return &ResponseMessage{
		Code:    200,
		Details: "successfully stored kv",
	}
}

func (n *NatsBackend) AddRecord(k KV) error {
	kv, err := n.JS.KeyValue(k.Bucket())
	if err != nil {
		return err
	}

	_, err = kv.Put(k.Key(), k.Value())
	if err != nil {
		return err
	}

	return nil
}

// getRecord wraps GetRecord by decrypting the returned value and handling resposnes.
func (n *NatsBackend) getRecord(k KV) *ResponseMessage {
	data, err := n.GetRecord(k)
	if err != nil && err == nats.ErrKeyNotFound {
		return &ResponseMessage{
			Code:  404,
			Error: "key not found",
		}
	}

	if err != nil && err != nats.ErrKeyNotFound {
		log.Println(err)
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	decrypted, err := decrypt(data, databaseKey)
	if err != nil {
		log.Println(err)
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	return &ResponseMessage{
		Details: string(decrypted),
		Code:    200,
	}

}

func (n *NatsBackend) GetRecord(k KV) ([]byte, error) {
	kv, err := n.JS.KeyValue(k.Bucket())
	if err != nil {
		return nil, err
	}

	v, err := kv.Get(k.Key())
	if err != nil {
		return nil, err
	}

	return v.Value(), nil
}

// deleteRecord wraps DeleteRecord and handles responses.
func (n *NatsBackend) deleteRecord(k KV) *ResponseMessage {
	err := n.DeleteRecord(k)
	if err != nil && err == nats.ErrKeyNotFound {
		return &ResponseMessage{
			Code:  404,
			Error: "key not found",
		}
	}

	if err != nil && err != nats.ErrKeyNotFound {
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	return &ResponseMessage{
		Code:    200,
		Details: "successfully deleted kv",
	}
}

func (n *NatsBackend) DeleteRecord(k KV) error {
	kv, err := n.JS.KeyValue(k.Bucket())
	if err != nil {
		return err
	}

	return kv.Delete(k.Key())
}

// unlock wraps Unlock and handles unmarshaling requests, verifying key size, and hanling responses
func (n *NatsBackend) unlock(m *nats.Msg) *ResponseMessage {
	var key DatabaseKey

	if databaseKey != nil {
		return &ResponseMessage{
			Code:  400,
			Error: "database already unlocked",
		}
	}

	if err := json.Unmarshal(m.Data, &key); err != nil {
		log.Printf("error unmarshaling json: %s", err)
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	if key.DBKey == "" || len(key.DBKey) < aes.BlockSize {
		return &ResponseMessage{
			Code:  400,
			Error: "key is too short",
		}
	}

	kv := NewJSRecord().SetBucket(piggyBucket).SetKey("init").SetValue(key.DBKey)

	if err := n.Unlock(kv); err != nil {
		log.Printf("error unlocking database: %s", err)
		return &ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
	}

	return &ResponseMessage{
		Code:    200,
		Details: "database successfully unlocked",
	}
}

func (n *NatsBackend) Unlock(k KV) error {
	key, err := fromBase64(string(k.Value()))
	if err != nil {
		return err
	}

	val, err := n.GetRecord(k)
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

// Lock locks the database by emptying the in memory key
func (n *NatsBackend) Lock() {
	databaseKey = nil
}

// Marshal marshals a ResponseMessage into a NATS message
func (r *ResponseMessage) Marshal() *nats.Msg {
	var data []byte
	var err error
	status := strconv.Itoa(r.Code)
	data, err = json.Marshal(r)
	if err != nil {
		data = []byte(`{"error": "internal server error"}`)
	}

	return &nats.Msg{
		Header: map[string][]string{
			"Status": []string{status},
		},
		Data: data,
	}
}
