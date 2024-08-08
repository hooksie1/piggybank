package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

var (
	testVals = map[string]string{
		"piggybank.secrets.secret1": "thesecret",
		"piggybank.secrets.secret2": "other secret",
		"piggybank.secrets.secret3": "this is another secret $@!)(*)/",
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

func setupEncryptedVals(t *testing.T, server *server.Server, vals map[string]string) ([]byte, AppContext) {

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
		KV:     kv,
		logger: logr.NewLogger(),
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

	for k, v := range vals {
		record := JetStreamRecord{
			encryptionKey: key,
			bucket:        piggyBucket,
			key:           k,
			value:         []byte(v),
		}
		if err := app.addRecord(&record); err != nil {
			t.Error(err)
		}
	}

	return key, app
}

func TestRotation(t *testing.T) {
	// reset key
	tt := []struct {
		name     string
		vals     map[string]string
		expected map[string]string
		rollback bool
		err      bool
	}{
		{
			name:     "normal rotation",
			rollback: false,
			vals:     testVals,
			expected: testVals,
			err:      false,
		},
		{
			name:     "rollback with error",
			rollback: true,
			vals:     testVals,
			expected: map[string]string{
				"piggybank.secrets.secret1": "thesecret",
				"piggybank.secrets.secret2": "other secret",
				"piggybank.secrets.secret3": "",
			},
			err: true,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			databaseKey = nil
			server := NewServer(t)
			defer shutdownJSServerAndRemoveStorage(t, server)

			key, app := setupEncryptedVals(t, server, v.vals)

			// Change one key with bad data to cause rollback
			if v.rollback {
				record := JetStreamRecord{
					encryptionKey: generateKey(),
					bucket:        piggyBucket,
					key:           "piggybank.secrets.secret3",
					value:         []byte("other secret"),
				}
				if err := app.addRecord(&record); err != nil {
					t.Error(err)
				}
			}

			_, err := app.Rotate(toBase64(key))
			if err != nil && !v.err {
				t.Fatal(err)
			}

			for sub, val := range v.vals {
				record := JetStreamRecord{
					bucket: piggyBucket,
					key:    sub,
				}
				decrypted, err := app.getRecord(&record, databaseKey)
				if err != nil && !v.err {
					t.Error(err)
				}

				if string(decrypted) != v.expected[sub] {
					t.Errorf("expected %s but got %s", val, string(decrypted))
				}
			}

		})
	}

}
