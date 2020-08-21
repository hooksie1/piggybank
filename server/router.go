package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ClientError interface {
	Error() string
	Body() ([]byte, error)
	Headers() (int, map[string]string)
}

type HTTPError struct {
	Cause   error  `json:"-"`
	Details string `json:"details"`
	Status  int    `json:"-"`
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Details
	}
	return e.Details + " : " + e.Cause.Error()
}

func (e *HTTPError) Body() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling response: %v", err)
	}
	return body, nil
}

func (e *HTTPError) Headers() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

func NewHTTPError(err error, status int, detail string) error {
	return &HTTPError{
		Cause:   err,
		Details: detail,
		Status:  status,
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Piggybank")
}

// authentication currently only supports basic auth.
func authentication(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()

		if user == "" || pass == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("A username and password must be supplied"))
			log.Printf("User credentials are empty")
			return
		}

		if !checkUser(user, pass) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s\n", "User or password incorrect")
			log.Printf("failed authentication attempt by %s", user)
			return
		}

		inner.ServeHTTP(w, r)

	})
}

// checkUser takes a user and password and compares that to
// the password stored in the boltDB database.
func checkUser(user, pass string) bool {

	dbUser := User{
		Username: user,
		Password: &Password{
			PlainText: pass,
		},
	}

	if err := dbUser.getUser(); err != nil {
		log.Printf("error checking user credentials: %s", err)
		return false
	}

	if dbUser.hash == "" {
		return false
	}

	valid := dbUser.compareHash()
	if !valid {
		return false
	}

	return true

}

// logger logs the endpoint requested and times how long the request takes.
func logger(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		user, _, _ := r.BasicAuth()

		log.Printf(
			"%s accessed %s %s %s",
			user,
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}

func checkDB(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := checkMasterPass(masterPass)
		if err != nil {
			w.WriteHeader(http.StatusPreconditionFailed)
			fmt.Fprintf(w, "{\"status\": \"locked\"}")
			return
		}

		inner.ServeHTTP(w, r)
	})
}

func (fn errHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err == nil {
		return
	}

	log.Printf("An error ocurred: %v", err)

	clientError, ok := err.(ClientError)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := clientError.Body()
	if err != nil {
		log.Printf("An error ocurred: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	status, headers := clientError.Headers()
	for k, v := range headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(status)
	w.Write(body)

}

func Backup(w http.ResponseWriter, r *http.Request) error {
	backupType := r.URL.Query().Get("type")

	if masterPass == nil {
		return NewHTTPError(nil, http.StatusPreconditionFailed, "You must unlock the database")
	}

	buser, _, _ := r.BasicAuth()
	if buser != "manager" {
		message := fmt.Sprintf("%s cannot retrieve secrets\n", buser)
		return NewHTTPError(nil, http.StatusUnauthorized, message)
	}
	if backupType == "local" {
		err := BackupLocal()
		if err != nil {
			return NewHTTPError(nil, http.StatusInternalServerError, "Error creating local backup")
		}
	}

	if backupType == "http" {
		BackupHTTP(w, r)
	}

	return nil

}

func (a *Application) Unmarshal(r io.Reader) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading request body: %s", err)
	}

	if err = json.Unmarshal(body, a); err != nil {
		return fmt.Errorf("error unmarshaling json data: %s", err)
	}

	return nil
}
