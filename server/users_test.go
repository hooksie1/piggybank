package server

import (
	"testing"
)

type testTable struct {
	expected string
	actual   string
}

func TestNewUser(t *testing.T) {
	user := newUser("testy")

	if user.Username != "testy" {
		t.Errorf("Username is incorrect. Expected testy, but got %s", user.Username)
	}

}

/* func TestAddUserRecord(t *testing.T) {
	os.Setenv("DATABASE_PATH", "./piggy.db")
	user := newUser("testy")
	user.Pass = NewPassword()

	err := user.addUser()
	if err != nil {
		t.Errorf("error adding user to database: %s", err)
	}

} */
