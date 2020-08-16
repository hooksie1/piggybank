package cmd

import (
	"io"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
)

type ClientError struct {
	Cause   error
	Details string
}

type RequestOption func(*http.Request) (*http.Request, error)

func (c *ClientError) Error() string {
	if c.Cause == nil {
		return c.Details
	}

	return c.Details + " : " + c.Cause.Error()
}

func NewRequest(opts ...RequestOption) (*http.Request, error) {
	r := &http.Request{}
	var err error

	for _, opt := range opts {
		r, err = opt(r)
		if err != nil {
			return nil, NewClientError(err, "Error creating request")
		}
	}

	return r, nil
}

func NewClientError(err error, detail string) error {
	return &ClientError{
		Cause:   err,
		Details: detail,
	}
}

func SetMethod(method string) RequestOption {
	return func(r *http.Request) (*http.Request, error) {
		r.Method = method
		return r, nil
	}
}

func SetURL(URL string) RequestOption {
	return func(r *http.Request) (*http.Request, error) {
		u, err := urlpkg.Parse(URL)
		if err != nil {
			return nil, err
		}
		r.URL = u
		return r, nil
	}
}

func SetQuery(query map[string]string) RequestOption {
	return func(r *http.Request) (*http.Request, error) {
		q := r.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}

		r.URL.RawQuery = q.Encode()

		return r, nil
	}
}

func SetBody(body io.Reader) RequestOption {
	return func(r *http.Request) (*http.Request, error) {
		rc, ok := body.(io.ReadCloser)
		if !ok && body != nil {
			rc = ioutil.NopCloser(body)
		}
		r.Body = rc
		return r, nil
	}
}

func SetCredentials(user, pass string) RequestOption {
	return func(r *http.Request) (*http.Request, error) {
		r.Header = make(http.Header)
		r.SetBasicAuth(user, pass)

		return r, nil
	}
}

func createQuery(r *http.Request, m map[string]string) *http.Request {
	query := r.URL.Query()

	for k, v := range m {
		query.Add(k, v)
	}

	r.URL.RawQuery = query.Encode()

	return r
}

func checkResponse(r *http.Response) error {

	if r.StatusCode == 401 {
		return NewClientError(nil, "Incorrect username/password")
	}

	if r.StatusCode == 412 {
		return NewClientError(nil, "You must unlock the database first")
	}

	if r.StatusCode < 200 || r.StatusCode > 299 {
		return NewClientError(nil, "Received a non 200 response")
	}

	return nil
}
