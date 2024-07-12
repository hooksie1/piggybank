package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

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

type ResponseError struct {
	Error string `json:"error"`
}

func NewDBRequest(verb DBVerb, key string) (Request, error) {
	subject, ok := SubjectVerbs[verb]
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

func NewRequest(verb Verb, key string) (Request, error) {
	subject := fmt.Sprintf("%s.%s", verb, key)
	return Request{
		Subject: subject,
		Data:    nil,
	}, nil
}

func (c *Client) Get(key string) (string, error) {
	subject := fmt.Sprintf("%s.%s.%s", secretSubject, GET, key)
	return c.Do(Request{Subject: subject, Data: nil})
}

func (c *Client) Post(key string, data []byte) (string, error) {
	subject := fmt.Sprintf("%s.%s.%s", secretSubject, POST, key)
	return c.Do(Request{Subject: subject, Data: data})
}

func (c *Client) Delete(key string) (string, error) {
	subject := fmt.Sprintf("%s.%s.%s", secretSubject, DELETE, key)
	return c.Do(Request{Subject: subject, Data: nil})
}

func (c *Client) Do(request Request) (string, error) {
	msg, err := c.Conn.Request(request.Subject, request.Data, 1*time.Second)
	if err != nil {
		return "", err
	}
	code := msg.Header.Get("Nats-Service-Error-Code")
	if code != "" {
		var respErr ResponseError
		fmt.Println(string(msg.Data))
		if err := json.Unmarshal(msg.Data, &respErr); err != nil {
			return "", err
		}
		return "", fmt.Errorf("status %s, details %v", code, respErr.Error)
	}

	return string(msg.Data), nil
}
