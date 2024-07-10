package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type DBVerb string

const (
	DBInit   DBVerb = "init"
	DBLock   DBVerb = "lock"
	DBUnlock DBVerb = "unlock"
	DBStatus DBVerb = "status"
)

var subjectVerbs = map[DBVerb]string{
	DBInit:   databaseInitSubject,
	DBLock:   databaseLockSubject,
	DBUnlock: databaseUnlockSubject,
	DBStatus: databaseStatusSubject,
}

type Client struct {
	Conn *nats.Conn
}

type DbRequest struct {
	Verb DBVerb
	Key  string
}

type Request struct {
	Subject string
	Data    []byte
}

func (d DBVerb) String() string {
	return string(d)
}

func GetClientDBVerbs() []string {
	return []string{DBInit.String(), DBLock.String(), DBUnlock.String(), DBStatus.String()}
}

func NewDBRequest(verb DBVerb, key string) (Request, error) {
	subject, ok := subjectVerbs[verb]
	if !ok {
		return Request{}, fmt.Errorf("invalid verb")
	}

	data, err := json.Marshal(DatabaseKey{DBKey: key})
	if err != nil {
		return Request{}, err
	}

	return Request{
		Subject: subject,
		Data:    data,
	}, nil
}

func (c *Client) Do(request Request) (string, error) {
	msg, err := c.Conn.Request(request.Subject, request.Data, 1*time.Second)
	if err != nil {
		return "", err
	}

	return string(msg.Data), nil
}
