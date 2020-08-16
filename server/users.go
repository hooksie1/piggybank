package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// User holds information about a user.
type User struct {
	Username string
	Pass     *Password
}

// NewUser returns a pointer to a new DBUser.
func newUser(username string) *User {
	pass := NewPassword()

	return &User{
		Username: username,
		Pass:     pass,
	}
}

func createUser(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	user := newUser(vars["userName"])

	buser, _, _ := r.BasicAuth()
	if buser != "manager" {
		message := fmt.Sprintf("%s cannot create users", buser)
		return NewHTTPError(nil, http.StatusUnauthorized, message)
	}

	if err := user.addUser(); err != nil {
		return fmt.Errorf("error adding user: %s", err)
	}

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error marshaling json data: %s", err)
	}

	fmt.Fprintf(w, string(data))

	return nil

}

func deleteUser(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	user := User{
		Username: vars["userName"],
	}

	buser, _, _ := r.BasicAuth()
	if buser != "manager" {
		message := fmt.Sprintf("%s cannot delete users", buser)
		return NewHTTPError(nil, http.StatusUnauthorized, message)
	}

	if err := user.deleteUser(); err != nil {
		return fmt.Errorf("error creating user: %s", err)
	}

	w.WriteHeader(http.StatusOK)

	return nil

}

func (u *User) checkManagerUser() bool {
	if u.Username != "manager" {
		return false
	}

	return true
}

func (u *User) addUser() error {
	record := &BoltRecord{
		Bucket: "Users",
		Key:    []byte(u.Username),
		Value:  []byte(u.Pass.hash),
	}

	if err := WriteRecord(record); err != nil {
		return fmt.Errorf("error adding user: %s", err)
	}

	return nil
}

func (u *User) deleteUser() error {
	record := &BoltRecord{
		Bucket: "Users",
		Key:    []byte(u.Username),
	}

	if err := record.DeleteRecord(); err != nil {
		return fmt.Errorf("error deleting user: %s", err)
	}

	return nil
}

func (u *User) getUser() error {
	record := &BoltRecord{
		Bucket: "Users",
		Key:    []byte(u.Username),
	}

	err := record.GetRecord()

	u.Pass.hash = string(record.Value)

	if err != nil {
		return fmt.Errorf("error getting user: %s", err)
	}

	return nil

}
