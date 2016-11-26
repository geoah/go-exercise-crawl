package crawler

import (
	"net/url"
	"strings"
	"sync"
)

// Result is a complete crawl of a site
type Result struct {
	sync.WaitGroup
	sync.RWMutex
	queue     chan *Target
	processed chan *Target
	targets   map[string]*Target
}

// StreamTargets returns targets as soon as they have been processed
func (r *Result) StreamTargets() chan *Target {
	return r.processed
}

// GetTargets is just a safe method to make sure the processing has finished
// before the targets can be returned.
// Since Crawl() doesn't wait for the result to be gathered, this does.
func (r *Result) GetTargets() map[string]*Target {
	r.Wait()
	return r.targets
}

// Enqueue cheks if we have already processed a URL and if not
// adds it to the queue to be processed by our workers.
func (r *Result) Enqueue(targetURL string) error {
	// validate that the given target is a URL we can use
	tURL, err := url.Parse(targetURL)
	if err != nil {
		return ErrInvalidURL
	}
	if tURL.Host == "" || tURL.Scheme == "" {
		return ErrInvalidURL
	}

	// normalize URL
	// TODO(geoah) Need better URL normalization
	tURL.Scheme = strings.ToLower(tURL.Scheme)
	tURL.Host = strings.ToLower(tURL.Host)
	tURL.Fragment = ""
	nURL := tURL.String()

	// check if we have already processed this target
	r.Lock()
	if _, exists := r.targets[nURL]; exists {
		r.Unlock()
		return nil
	}
	// if not, create a new target
	target := &Target{
		url:       tURL,
		assetURLs: map[string]int{},
		linkURLs:  map[string]int{},
	}
	r.targets[nURL] = target

	// and add it to the queue
	r.Add(1)
	r.queue <- target
	r.Unlock()
	return nil
}
