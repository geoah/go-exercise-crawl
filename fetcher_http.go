package crawler

import (
	"io"
	"net/http"
)

const (
	defaultMaxRetries = 5
)

// FetcherHTTP is an implementation of Fetcher for HTTP
type FetcherHTTP struct {
	client     *http.Client
	maxRetries int
}

// NewFetcherHTTP creates a new FetcherHTTP
func NewFetcherHTTP(client *http.Client, maxRetries int) *FetcherHTTP {
	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}
	return &FetcherHTTP{client, maxRetries}
}

// Fetch will GET a URL and return it's body reader
func (f *FetcherHTTP) Fetch(target *Target) (io.Reader, error) {
	// bump tries and remove error
	target.tries++
	target.err = nil

	shouldRetry := false
	if target.tries < f.maxRetries {
		shouldRetry = true
	}

	// construct and make our HTTP request
	req, err := http.NewRequest("GET", target.url.String(), nil)
	if err != nil {
		target.err = err
		return nil, FetcherError{err, shouldRetry}
	}
	resp, err := f.client.Do(req)
	if err != nil {
		target.err = err
		return nil, FetcherError{err, shouldRetry}
	}

	// TODO(geoah) Handle other HTTP codes a bit more generically; eg 2xx, 4xx, 5xx
	// TODO(geoah) Handling 429 (too-many-requests) errors specially would also be nice
	// figure out if we need to parse the page or return an error
	switch resp.StatusCode {
	case http.StatusNoContent:
		// if there is no content return an empty reader
		return nil, nil
	case http.StatusNotFound,
		http.StatusUnauthorized,
		http.StatusBadRequest:
		// don't retry and just return an error
		// TODO(geoah) Maybe return a new ErrPageNotExist error
		target.tries = f.maxRetries + 1
		target.err = ErrFetchingPage
		return nil, FetcherError{target.err, false}
	case http.StatusInternalServerError,
		http.StatusTooManyRequests:
		// return error but allow retry
		target.err = ErrFetchingPage
		return nil, FetcherError{target.err, shouldRetry}
	}

	// since we managed to get this far, clean the error
	target.err = nil
	return resp.Body, nil
}
