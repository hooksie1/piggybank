package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ClientError represents a non-server error
type ClientError struct {
	// Status is the status code to be returned
	Status int

	// Details are a nicely formatted client error
	Details string

	//Objects is a slice of error objects
	ErrorsWithMetadata []ErrorWithMetadata

	//DetailedError is the actual error to be logged
	DetailedError error

	// Additional parameters for the error
	Params map[string]any
}

type ErrorWithMetadata struct {
	// Code assigned to this type of error
	Code string `json:"code"`

	// Message is the details for the error
	Message string `json:"message"`

	// Type is the type of error
	Type string `json:"type"`

	// Level is the level of the error
	Level string `json:"level"`
}

type ClientErrorOpt func(*ClientError)

func (c ClientError) Error() string {
	return c.Details
}

// Body formats the error as JSON to return to the client.
func (c ClientError) Body() []byte {
	format := `{"code": %q, "message": %q, "type": %q, "level": %q}`
	var baseJSON string
	if c.ErrorsWithMetadata != nil {
		baseJSON = fmt.Sprintf(`{"errors": [%s]`, errorMetadataToFormattedString(format, c.ErrorsWithMetadata...))
	} else {
		baseJSON = fmt.Sprintf(`{"errors": [%q]`, c.Details)
	}

	if len(c.Params) > 0 {
		paramsJSON, _ := json.Marshal(c.Params)
		baseJSON = fmt.Sprintf("%s, \"params\": %s", baseJSON, string(paramsJSON))
	}

	return []byte(baseJSON + "}")
}

func (c ClientError) Code() int {
	return c.Status
}

// LoggedError is to be logged when returning a client error
func (c ClientError) LoggedError() string {
	format := `{code: %s, message: %s, type: %s, level: %s}`
	if c.ErrorsWithMetadata != nil {
		return fmt.Sprintf("%s %s", c.Details, errorMetadataToFormattedString(format, c.ErrorsWithMetadata...))
	}

	return fmt.Sprintf("%s %s", c.Details, c.DetailedError.Error())
}

// errorMetadataToFormattedString takes a format string and a slice of ErrorWithMetadata objects and returns a formatted string
func errorMetadataToFormattedString(format string, objects ...ErrorWithMetadata) string {
	var errors []string
	for _, v := range objects {
		errors = append(errors, fmt.Sprintf(format, v.Code, v.Message, v.Type, v.Level))
	}

	return strings.Join(errors, ", ")
}

func (c ClientError) As(target any) bool {
	_, ok := target.(*ClientError)
	return ok
}

// WithMetadataErrors is a variadic function that takes Errors With Metadata for returning error objects to clients
func (c ClientError) WithMetadataErrors(objects ...ErrorWithMetadata) ClientError {
	c.ErrorsWithMetadata = objects

	return c
}

func WithDetailedError(err error) ClientErrorOpt {
	return func(c *ClientError) {
		c.DetailedError = err
	}
}

func WithAdditionalParams(params map[string]any) ClientErrorOpt {
	return func(c *ClientError) {
		if len(params) == 0 {
			return
		}
		c.Params = params
	}
}

// NewClientError returns a new client error. If no metadata errors are given, default error metadata is returned
func NewClientError(err error, code int, opts ...ClientErrorOpt) ClientError {
	metadata := ErrorWithMetadata{
		Code:    "CWGEN1",
		Message: err.Error(),
		Type:    "api",
		Level:   "warning",
	}
	ce := ClientError{
		Status:             code,
		Details:            err.Error(),
		ErrorsWithMetadata: []ErrorWithMetadata{metadata},
		DetailedError:      err,
	}

	for _, v := range opts {
		v(&ce)
	}

	return ce
}
