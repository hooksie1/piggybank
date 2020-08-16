package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	PlainText string `json:"password"`
	hash      string
}

type Application struct {
	Application   string `json:"application"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	encryptedPass []byte
}

type MasterPass struct {
	MasterPassword string `json:"master_password"`
}

// NewPassword returns a pointer to a new password.
func NewPassword() *Password {
	pass := generatePass()
	return &Password{
		PlainText: pass,
		hash:      HashPassword(pass),
	}
}

func generatePass() string {
	pass := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, pass)
	if err != nil {
		panic(err)
	}

	secret := base64.RawStdEncoding.EncodeToString(pass)

	return string(secret)
}

func HashPassword(s string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}

func (u *User) compareHash() bool {

	err := bcrypt.CompareHashAndPassword([]byte(u.Pass.hash), []byte(u.Pass.PlainText))
	if err != nil {
		log.Printf("failed password check: %s", err)
		return false
	}

	return true

}

func addPass(w http.ResponseWriter, r *http.Request) error {
	buser, _, _ := r.BasicAuth()
	if buser == "manager" {
		message := fmt.Sprintf("%s cannot create secrets", buser)
		return NewHTTPError(nil, http.StatusUnauthorized, message)
	}

	if err := checkLength(); err != nil {
		return NewHTTPError(err, http.StatusPreconditionFailed, "database not unlocked")
	}

	var application Application

	if err := application.Unmarshal(r.Body); err != nil {
		return err
	}

	bytePass := []byte(application.Password)

	encryptedPass, err := encrypt(bytePass, []byte(masterPass))
	if err != nil {
		return fmt.Errorf("error encrypting password: %s", err)
	}

	application.encryptedPass = encryptedPass

	record, err := application.generateRecord()
	if err != nil {
		return fmt.Errorf("error generating record: %s", err)
	}

	if err := WriteRecord(record); err != nil {
		return err
	}

	return nil

}

// generateRecord creates a record used in the BoltDB database out
// using data from an application.
func (a *Application) generateRecord() (*BoltRecord, error) {
	encryptedUser, err := encrypt([]byte(a.Username), masterPass)
	if err != nil {
		return nil, err
	}
	return &BoltRecord{
		Bucket: a.Application,
		Key:    encryptedUser,
		Value:  a.encryptedPass,
	}, nil

}

// generateApplication creates an application from the data in
// a BoltDB record.
func (r *BoltRecord) generateApplication() (*Application, error) {
	decryptedPass, err := decrypt(r.Value, masterPass)
	if err != nil {
		return nil, err
	}

	decryptedUser, err := decrypt(r.Key, masterPass)
	if err != nil {
		return nil, err
	}

	return &Application{
		Application: r.Bucket,
		Username:    string(decryptedUser),
		Password:    string(decryptedPass),
	}, nil
}

func getPass(w http.ResponseWriter, r *http.Request) error {
	app := r.URL.Query().Get("application")
	user := r.URL.Query().Get("username")

	buser, _, _ := r.BasicAuth()
	if buser == "manager" {
		return NewHTTPError(nil, http.StatusUnauthorized, "manager account not allowed to view credentials")
	}

	record, err := NewBoltRecord(
		BoltBucket(app),
		BoltKey([]byte(user)),
	)
	if err != nil {
		return fmt.Errorf("Error getting value for record: %s", err)
	}

	if err := record.GetRecord(); err != nil {
		return fmt.Errorf("error getting record: %s", err)
	}

	application, err := record.generateApplication()
	if err != nil {
		return fmt.Errorf("error generating application: %s", err)
	}

	data, err := json.Marshal(application)
	if err != nil {
		return fmt.Errorf("error unmarshaling json data: %s", err)
	}

	w.Write(data)

	return nil

}

func deletePass(w http.ResponseWriter, r *http.Request) error {
	app := r.URL.Query().Get("application")
	user := r.URL.Query().Get("username")

	buser, _, _ := r.BasicAuth()
	if buser == "manager" {
		return NewHTTPError(nil, http.StatusUnauthorized, "manager account not allowed to delete credentials")
	}

	record := &BoltRecord{
		Bucket: app,
		Key:    []byte(user),
	}

	if err := record.DeleteRecord(); err != nil {
		return fmt.Errorf("error deleting record: %s", err)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}
