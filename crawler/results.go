package crawler

import (
	"sort"
	"sync"
)

// Results from a complete crawl of a site
type Results struct {
	sync.WaitGroup
	sync.RWMutex
	queue   chan string
	results map[string]*Result
}

// GetResults is just a safe method to make sure the processing has finished
// before the results can be returned.
// Since Crawl() doesn't wait for the results to be gathered, this does.
func (r *Results) GetResults() map[string]*Result {
	r.Wait()
	return r.results
}

// Enqueue cheks if we have already processed a URL and if not
// adds it to the queue to be processed by our workers.
func (r *Results) Enqueue(target string) {
	r.Lock()
	if _, exists := r.results[target]; exists {
		r.Unlock()
		return
	}
	r.results[target] = nil
	r.Add(1)
	r.queue <- target
	r.Unlock()
}

// Result from each page of a crawl
// Assets and Links are represented as maps using their URL for keys
// and the times they appear in the page as value.
type Result struct {
	AssetURLs map[string]int
	LinkURLs  map[string]int
}

// GetAssetURLs -
func (r *Result) GetAssetURLs(sorted bool) []string {
	urls := make([]string, len(r.AssetURLs))
	i := 0
	for url := range r.AssetURLs {
		urls[i] = url
		i++
	}
	if sorted {
		sort.Strings(urls)
	}
	return urls
}

// GetLinkURLs -
func (r *Result) GetLinkURLs(sorted bool) []string {
	urls := make([]string, len(r.LinkURLs))
	i := 0
	for url := range r.LinkURLs {
		urls[i] = url
		i++
	}
	if sorted {
		sort.Strings(urls)
	}
	return urls
}
