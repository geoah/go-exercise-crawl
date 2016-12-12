package crawler

import (
	"net/url"
	"sort"
	"strings"
)

// Target is a single page (URL) we are trying to retrieve
// In case we weren't able to fetch or parse the page, an error will be present
// Assets and Links are represented as maps using their URL for keys
// and the times they appear in the page as value.
type Target struct {
	url   *url.URL
	tries int
	err   error
	done  bool

	assetURLs map[string]int
	linkURLs  map[string]int
}

// NewTarget -
func NewTarget(targetURL string) (*Target, error) {
	// validate that the given target is a URL we can use
	tURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, ErrInvalidURL
	}
	if tURL.Host == "" || tURL.Scheme == "" {
		return nil, ErrInvalidURL
	}

	// normalize URL
	// TODO(geoah) Need better URL normalization
	tURL.Scheme = strings.ToLower(tURL.Scheme)
	tURL.Host = strings.ToLower(tURL.Host)
	tURL.Fragment = ""

	tgt := &Target{
		url:       tURL,
		assetURLs: map[string]int{},
		linkURLs:  map[string]int{},
	}

	return tgt, nil
}

// String -
func (t *Target) String() string {
	return t.url.String()
}

// GetURL returns URL
func (t *Target) GetURL() *url.URL {
	return t.url
}

// GetError return error
func (t *Target) GetError() error {
	return t.err
}

// GetAssetURLs return asset URLs
func (t *Target) GetAssetURLs(sorted bool) []string {
	urls := make([]string, len(t.assetURLs))
	i := 0
	for url := range t.assetURLs {
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
	urls := make([]string, len(t.linkURLs))
	i := 0
	for url := range t.linkURLs {
		urls[i] = url
		i++
	}
	if sorted {
		sort.Strings(urls)
	}
	return urls
}
