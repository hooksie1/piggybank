package server

import (
	"crypto/aes"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

var (
	ErrInitialized = errors.New("database already initialized")
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

func (n *NatsBackend) SetupMicro() error {
	log.Println("setting up micro")
	srv, err := micro.AddService(n.Conn, micro.Config{
		Name:    "Piggybank",
		Version: "1.0.0",
		Endpoint: &micro.EndpointConfig{
			Subject: "piggybank.>",
			Handler: micro.HandlerFunc(n.HandleRequests),
		},
	})
	if err != nil {
		return err
	}

	databaseGroup := srv.AddGroup("piggybank.database")
	if err := databaseGroup.AddEndpoint("lock", micro.HandlerFunc(n.LockRequest)); err != nil {
		return err
	}

	if err := databaseGroup.AddEndpoint("unlock", micro.HandlerFunc(n.UnlockRequest)); err != nil {
		return err
	}

	if err := databaseGroup.AddEndpoint("initialize", micro.HandlerFunc(n.InitializeRequest)); err != nil {
		return err
	}

	return nil
}

// HandleRequests handles a non database specific request
func (n *NatsBackend) HandleRequests(req micro.Request) {

	if databaseKey == nil {
		msg := ResponseMessage{
			Code:  403,
			Error: "database locked",
		}
		req.Respond(msg.body(), micro.WithHeaders(msg.headers()))
	}

	msg := n.HandleKeyAction(req)
	req.Respond(msg.body(), micro.WithHeaders(msg.headers()))
}

// HandleKeyAction handles the action for the requested key based on the method in the header
func (n *NatsBackend) HandleKeyAction(req micro.Request) *ResponseMessage {
	var msg *ResponseMessage
	record := NewJSRecord().SetBucket(piggyBucket).SetSanitizedKey(req.Subject())

	switch req.Headers().Get("method") {
	case "post":
		record.SetValue(string(req.Data()))
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

// LockRequest locks the database
func (n *NatsBackend) LockRequest(req micro.Request) {
	n.Lock()

	m := ResponseMessage{
		Code:    200,
		Details: "database locked",
	}

	req.Respond(m.body(), micro.WithHeaders(m.headers()))
}

// UnlockRequest unlocks the database if it is locked
func (n *NatsBackend) UnlockRequest(req micro.Request) {
	var unlocked bool
	kv := NewJSRecord().SetBucket(piggyBucket).SetKey("init")
	if databaseKey != nil {
		unlocked = true
	}

	if unlocked {
		m := ResponseMessage{
			Code:  400,
			Error: "database already unlocked",
		}
		req.Respond(m.body(), micro.WithHeaders(m.headers()))
	}

	_, err := n.GetRecord(kv)
	if err != nil && err != nats.ErrKeyNotFound {
		log.Println(err)
		m := ResponseMessage{
			Code:  500,
			Error: "internal server error",
		}
		req.Respond(m.body(), micro.WithHeaders(m.headers()))
	}

	if err == nats.ErrKeyNotFound {
		m := ResponseMessage{
			Code:  400,
			Error: "database not initialized",
		}
		req.Respond(m.body(), micro.WithHeaders(m.headers()))
	}

	m := n.unlock(req.Data())
	req.Respond(m.body(), micro.WithHeaders(m.headers()))
}

// InitializeRequest intitializes the database if it is uninitialized
func (n *NatsBackend) InitializeRequest(req micro.Request) {
	m := n.initialize()
	req.Respond(m.body(), micro.WithHeaders(m.headers()))
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
func (n *NatsBackend) unlock(data []byte) *ResponseMessage {
	var key DatabaseKey

	if databaseKey != nil {
		return &ResponseMessage{
			Code:  400,
			Error: "database already unlocked",
		}
	}

	if err := json.Unmarshal(data, &key); err != nil {
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

func (r *ResponseMessage) body() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		return []byte(`{"error": "internal server error"}`)
	}

	return data
}

func (r *ResponseMessage) headers() map[string][]string {
	return map[string][]string{
		"Status": {strconv.Itoa(r.Code)},
	}
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
			"Status": {status},
		},
		Data: data,
	}
}
