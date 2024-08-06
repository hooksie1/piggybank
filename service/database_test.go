package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

var (
	testVals = map[string]string{
		"piggybank.secrets.secret1": "thesecret",
		"piggybank.secrets.secret2": "other secret",
	}
)

func NewServer(t *testing.T) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.JetStream = true
	opts.StoreDir = t.TempDir()
	opts.Port = -1
	return natsserver.RunServer(&opts)
}

func shutdownJSServerAndRemoveStorage(t *testing.T, s *server.Server) {
	t.Helper()
	var sd string
	if config := s.JetStreamConfig(); config != nil {
		sd = config.StoreDir
	}
	s.Shutdown()
	if sd != "" {
		if err := os.RemoveAll(sd); err != nil {
			t.Fatalf("Unable to remove storage %q: %v", sd, err)
		}
	}
	s.WaitForShutdown()
}

func TestRotate(t *testing.T) {
	server := NewServer(t)

	// nats connection
	nc, err := nats.Connect(server.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	// jetstream
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	// creates the bucket
	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "piggybank"})
	if err != nil {
		t.Fatal(err)
	}

	app := AppContext{
		KV: kv,
	}

	key, err := app.initialize()
	if err != nil {
		t.Fatal(err)
	}

	dbKey := DatabaseKey{DBKey: toBase64(key)}
	data, err := json.Marshal(dbKey)
	if err != nil {
		t.Fatal(err)
	}

	if err := app.unlock(data); err != nil {
		t.Fatal(err)
	}

	for k, v := range testVals {
		record := NewJSRecord().SetEncryptionKey(key).SetBucket(piggyBucket).SetKey(k).SetValue(v)
		if err := app.addRecord(record); err != nil {
			t.Error(err)
		}
	}

	_, err = app.Rotate(string(key))
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range testVals {
		record := NewJSRecord().SetBucket(piggyBucket).SetKey(k)
		decrypted, err := app.getRecord(record)
		if err != nil {
			t.Error(err)
		}

		if string(decrypted) != v {
			t.Errorf("expected %s but got %s", v, string(decrypted))
		}
	}
}
