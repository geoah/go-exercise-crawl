package crawler

import "io"

// Fetcher attemps to fetch the content of a Target and return
// a reader for it, or an error
type Fetcher interface {
	Fetch(target *Target) (io.Reader, error)
}

// FetcherError is a custom error to allow us to retry
type FetcherError struct {
	error
	shouldRetry bool
}

// ShouldRetry wether we should retry a failed request
func (e *FetcherError) ShouldRetry() bool {
	return e.shouldRetry
}
