package crawler

import (
	"net/url"
	"sort"
)

// Target is each of the URLs we are trying to retrieve
// In case we weren't able to fetch or parse the page, an error will be present
// Assets and Links are represented as maps using their URL for keys
// and the times they appear in the page as value.
type Target struct {
	url   *url.URL
	tries int
	err   error

	AssetURLs map[string]int
	LinkURLs  map[string]int
}

// GetError return error
func (t *Target) GetError() error {
	return t.err
}

// GetAssetURLs return asset URLs
func (t *Target) GetAssetURLs(sorted bool) []string {
	urls := make([]string, len(t.AssetURLs))
	i := 0
	for url := range t.AssetURLs {
		urls[i] = url
		i++
	}
	if sorted {
		sort.Strings(urls)
	}
	return urls
}

// GetLinkURLs return link URLs
func (t *Target) GetLinkURLs(sorted bool) []string {
	urls := make([]string, len(t.LinkURLs))
	i := 0
	for url := range t.LinkURLs {
		urls[i] = url
		i++
	}
	if sorted {
		sort.Strings(urls)
	}
	return urls
}
