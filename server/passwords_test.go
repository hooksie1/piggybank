package server

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type TB struct {
	name     string
	expected string
	actual   string
}

func TestGeneratePass(t *testing.T) {
	secret := generatePass()
	secret2 := generatePass()

	if len(secret) != 43 {
		t.Errorf("Secret is not 32 bytes")
	}

	if secret == secret2 {
		t.Errorf("Secrets are the same.")
	}

}

func TestHashPassword(t *testing.T) {
	plainText := "this is the password"
	pass := HashPassword(plainText)

	if err := bcrypt.CompareHashAndPassword([]byte(pass), []byte(plainText)); err != nil {
		t.Errorf("Password does not match hash")
	}
}

func TestNewPassword(t *testing.T) {
	pass := NewPassword()

	plainText := pass.PlainText
	hash := pass.hash

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainText)); err != nil {
		t.Errorf("Password does not match hash.")
	}
}

func TestCompareHash(t *testing.T) {
	user := newUser("testing")

	ok := user.compareHash()
	if !ok {
		t.Errorf("Password %s does not match hash %s.", user.Pass.PlainText, user.Pass.hash)
	}

}

func TestGenerateRecord(t *testing.T) {
	masterPass = generateKey()

	encPass, err := encrypt([]byte("testPass"), masterPass)
	if err != nil {
		t.Errorf("Error encrypting password")
	}

	app := Application{
		Application:   "testApp",
		Username:      "testUser",
		encryptedPass: encPass,
		Password:      "testPass",
	}

	record, err := app.generateRecord()
	if err != nil {
		t.Errorf("Error in generating record")
	}

	dUser, err := decrypt(record.Key, masterPass)
	if err != nil {
		t.Errorf("Error decrypting user")
	}

	dPass, err := decrypt(record.Value, masterPass)
	if err != nil {
		t.Errorf("Error decrypting password")
	}

	tt := []TB{
		{name: "app", expected: app.Application, actual: record.Bucket},
		{name: "user", expected: app.Username, actual: string(dUser)},
		{name: "pass", expected: app.Password, actual: string(dPass)},
	}

	for _, v := range tt {
		if v.actual != v.expected {
			t.Errorf("For %s: expected %v, but got %v", v.name, v.expected, v.actual)
		}
	}

}
