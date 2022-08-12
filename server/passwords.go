package server

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	PlainText string `json:"password"`
	hash      string
}

type DatabaseKey struct {
	DBKey string `json:"database_key"`
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
