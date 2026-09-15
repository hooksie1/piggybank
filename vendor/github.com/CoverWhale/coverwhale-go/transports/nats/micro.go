package nats

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cwerrors "github.com/CoverWhale/coverwhale-go/errors"
	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/segmentio/ksuid"
)

type HandlerWithErrors func(*logr.Logger, micro.Request) error

// Interface of a type of client error, not necessarily the cwerrors.ClientError type
type ClientError interface {
	Error() string
	Code() int
	Body() []byte
	LoggedError() string
}

func HandleNotify(s micro.Service, healthFuncs ...func(chan<- string, micro.Service)) error {
	stopChan := make(chan string, 1)
	for _, v := range healthFuncs {
		go v(stopChan, s)
	}

	go handleNotify(stopChan)

	logr.Info(<-stopChan)
	return s.Stop()
}

func handleNotify(stopChan chan<- string) {
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigTerm
	stopChan <- fmt.Sprintf("received signal: %v", sig)
}

// ErrorHandler wraps a normal micro endpoint and allows for returning errors natively. Errors are
// checked and if an error is a client error, details are returned, otherwise a 500 is returned and logged
func ErrorHandler(logger *logr.Logger, h HandlerWithErrors) micro.HandlerFunc {
	return func(r micro.Request) {
		start := time.Now()
		id, err := SubjectToRequestID(r.Subject())
		if err != nil {
			handleRequestError(logger, cwerrors.NewClientError(err, 400), r)
			return
		}
		reqLogger := logger.WithContext(map[string]string{"request_id": id, "path": r.Subject()})
		defer func() {
			reqLogger.Infof("duration %dms", time.Since(start).Milliseconds())
		}()

		correlationId := r.Headers().Get("X-Correlation-Id")
		if correlationId != "" {
			reqLogger = reqLogger.WithContext(map[string]string{"correlation_id": correlationId})
		}

		if err := buildQueryHeaders(r); err != nil {
			handleRequestError(reqLogger, err, r)
		}

		if r.Headers().Get("X-Request-ID") == "" && len(r.Headers()) != 0 {
			r.Headers()["X-Request-ID"] = []string{id}
		}

		err = h(reqLogger, r)
		if err == nil {
			return
		}

		handleRequestError(reqLogger, err, r)
	}
}

// Create CW specific headers from the NATS bridge plugin headers
func buildQueryHeaders(r micro.Request) error {
	headers := nats.Header(r.Headers())
	query := headers.Get("X-NatsBridge-UrlQuery")
	parsed, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	for k, v := range parsed {
		key := fmt.Sprintf("X-CW-%s", k)
		headers[key] = v
	}

	return nil
}

func GetQueryHeaders(headers micro.Headers, key string) []string {
	k := fmt.Sprintf("X-CW-%s", key)
	return headers.Values(k)
}

// handleRequestError will return a client error if it is a client error, otherwise it will return a 500
func handleRequestError(logger *logr.Logger, err error, r micro.Request) {
	ce, ok := err.(ClientError)
	if ok {
		logger.Error(ce.LoggedError())
		r.Error(fmt.Sprintf("%d", ce.Code()), http.StatusText(ce.Code()), ce.Body())
	}

	logger.Error(err.Error())

	r.Error("500", "internal server error", []byte(`{"errors": [{"code": "CWINT1", "message": "internal server error", "type": "server", "level": "warning"}]}`))
}

func SubjectToRequestID(s string) (string, error) {
	split := strings.Split(s, ".")
	if len(split) < 3 {
		return "", fmt.Errorf("invalid subject")
	}

	id := split[3]

	_, err := ksuid.Parse(id)
	if err != nil {
		return "", fmt.Errorf("invalid ksuid request ID")
	}

	return id, nil
}

func RequestLogger(l *logr.Logger, subject string) (*logr.Logger, error) {
	id, err := SubjectToRequestID(subject)
	if err != nil {
		return nil, err
	}
	return l.WithContext(map[string]string{"request_id": id}), nil
}
