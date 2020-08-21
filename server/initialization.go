package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go.etcd.io/bbolt"
)

type Status struct {
	State string `json:"state"`
}

// initializeManager creates the manager user and returns the
// plaintext password for the manager user.
func initializeManager() (string, error) {
	user := newUser("manager")

	err := user.addUser()
	if err != nil {
		return "", err
	}

	return user.PlainText, nil
}

func initializeDB(key []byte) error {
	random := generatePass()

	encryptedString, err := encrypt([]byte(random), key)
	if err != nil {
		return fmt.Errorf("error encrypting initial password: %s", err)
	}

	br, err := NewBoltRecord(
		BoltBucket("initialized"),
		BoltKey([]byte("init")),
		BoltValue([]byte(encryptedString)),
	)
	if err != nil {
		return err
	}

	if err = WriteRecord(br); err != nil {
		return err
	}

	return nil

}

func initialize(w http.ResponseWriter, r *http.Request) error {

	status, err := checkInitStatus()
	if err != nil {
		return fmt.Errorf("error checking status: %s", err)
	}

	if status {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{\"status\": \"initialized\"}")
		return nil
	}

	key := generateKey()

	if err = initializeDB(key); err != nil {
		return fmt.Errorf("error initializing database: %s", err)
	}

	managerPass, err := initializeManager()
	if err != nil {
		return fmt.Errorf("error initializing manager account: %s", err)
	}

	secret := toBase64(key)

	fmt.Fprintf(w, "Master decrypt password is %s\nUser manager username: manager password: %s\n", secret, managerPass)

	return nil
}

func getStatus(w http.ResponseWriter, r *http.Request) error {
	status, err := checkInitStatus()
	if err != nil {
		return fmt.Errorf("error checking status: %s", err)
	}

	if !status {
		return NewHTTPError(err, http.StatusPreconditionFailed, "status: uninitialized")
	}

	fmt.Fprintf(w, "{\"status\": \"initialized\"}")

	return nil

}

func checkInitStatus() (bool, error) {

	status := false

	_, err := os.Stat(dbName)
	if os.IsNotExist(err) {
		return status, nil
	}

	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Printf("error opening database: %s", err)
		return status, err
	}

	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("initialized"))
		if err != nil {
			return err
		}

		if b.Stats().KeyN == 0 {
			err = fmt.Errorf("Application not initialized")
			return err
		}

		status = true
		return nil
	})
	if err != nil {
		return false, err
	}

	return status, nil
}
