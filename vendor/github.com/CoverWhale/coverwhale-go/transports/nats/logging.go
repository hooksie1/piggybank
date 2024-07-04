package nats

import (
	"github.com/nats-io/nats.go"
)

type NatsLogger struct {
	subject string
	conn    *nats.Conn
}

func NewNatsLogger(subject string, nc *nats.Conn) NatsLogger {
	return NatsLogger{
		subject: subject,
		conn:    nc,
	}
}

func (n NatsLogger) Write(p []byte) (int, error) {
	err := n.conn.Publish(n.subject, p)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
