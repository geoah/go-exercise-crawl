package crawler

import (
	"net/url"
	"strings"
	"sync"
)

// Results from a complete crawl of a site
type Results struct {
	sync.WaitGroup
	sync.RWMutex
	queue   chan *Target
	results map[string]*Target
}

// GetTargets is just a safe method to make sure the processing has finished
// before the targets can be returned.
// Since Crawl() doesn't wait for the results to be gathered, this does.
func (r *Results) GetTargets() map[string]*Target {
	r.Wait()
	return r.results
}

// Enqueue cheks if we have already processed a URL and if not
// adds it to the queue to be processed by our workers.
func (r *Results) Enqueue(targetURL string) error {
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
	if _, exists := r.results[nURL]; exists {
		r.Unlock()
		return nil
	}
	// if not, create a new target
	target := &Target{
		url:       tURL,
		AssetURLs: map[string]int{},
		LinkURLs:  map[string]int{},
	}
	r.results[nURL] = target

	// and add it to the queue
	r.Add(1)
	r.queue <- target
	r.Unlock()
	return nil
}
