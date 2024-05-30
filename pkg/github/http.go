package github

import (
	"errors"
	"net/http"
)

// Transport is an http.RoundTripper that authenticates all requests
// using the provided TokenSource.
type Transport struct {
	Source TokenSource
	Base   http.RoundTripper
}

// RoundTrip executes a single HTTP transaction.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Source == nil {
		return nil, errors.New("no token source provided")
	}

	token, err := t.Source.Token()
	if err != nil {
		return nil, err
	}

	token.SetAuthHeader(req)

	return t.base().RoundTrip(req)
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}
