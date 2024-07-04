package service

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go/micro"
	"github.com/segmentio/ksuid"
)

const (
	Bucket = "piggybank"
)

type AppHandlerFunc func(micro.Request, AppContext) error

type ClientError struct {
	Code    int
	Details string
}

func (c ClientError) Error() string {
	return c.Details
}

func (c *ClientError) Body() []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"}`, c.Details))
}

func (c *ClientError) CodeString() string {
	return strconv.Itoa(c.Code)
}

func (c ClientError) As(target any) bool {
	_, ok := target.(*ClientError)
	return ok
}

func NewClientError(err error, code int) ClientError {
	return ClientError{
		Code:    code,
		Details: err.Error(),
	}
}

// ErrorHandler wraps a normal micro endpoint and allows for returning errors natively. Errors are
// checked and if an error is a client error, details are returned, otherwise a 500 is returned and logged
func AppHandler(logger *logr.Logger, h AppHandlerFunc, app AppContext) micro.HandlerFunc {
	return func(r micro.Request) {
		start := time.Now()
		id := ksuid.New().String()
		reqLogger := logger.WithContext(map[string]string{"request_id": id, "path": r.Subject()})
		defer func() {
			reqLogger.Infof("duration %dms", time.Since(start).Milliseconds())
		}()

		app.logger = reqLogger

		err := h(r, app)
		if err == nil {
			return
		}

		handleRequestError(reqLogger, err, r)
	}
}

func handleRequestError(logger *logr.Logger, err error, r micro.Request) {
	var ce ClientError
	if errors.As(err, &ce) {
		r.Error(ce.CodeString(), http.StatusText(ce.Code), ce.Body())
		return
	}

	logger.Error(err)

	r.Error("500", "internal server error", []byte(`{"error": "internal server error"}`))
}
