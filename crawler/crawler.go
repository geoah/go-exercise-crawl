package crawler

import (
	"errors"
	"net/http"
	"net/url"
	"sort"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	// ErrInvalidURL trying to crawl an invalid URL
	ErrInvalidURL = errors.New("Invalid URL")
)

// Crawler is basically an HTTP scraper
type Crawler struct {
	client        *http.Client
	crawlExternal bool
}

// New creates a new Crawler from an HTTP client
func New(client *http.Client, crawlExternal bool) *Crawler {
	return &Crawler{client, crawlExternal}
}

// Result results from a crawl
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

// Crawl -
func (c *Crawler) Crawl(target string) (*Result, error) {
	// validate that the given target is actually a URL
	tURL, err := url.Parse(target)
	if err != nil {
		return nil, ErrInvalidURL
	}

	// construct and make our HTTP request
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	// make sure to close body when all is done
	defer resp.Body.Close()

	// find out our location from headers as we might have been redirected
	loc, err := resp.Location()
	if err != nil {
		// if we can't we can always fall back to the original target
		loc = tURL
	}

	// init our result
	result := &Result{
		AssetURLs: map[string]int{},
		LinkURLs:  map[string]int{},
	}

	// and we can now start going through our response body
	doc := html.NewTokenizer(resp.Body)
	tokenType := doc.Next()
	for tokenType != html.ErrorToken {
		token := doc.Token()
		// we only care about starting elements
		if tokenType == html.StartTagToken {
			switch token.DataAtom {
			case atom.A:
				// we assume that <a> elements will provide us with all links
				// go through the attributes and find the HREF one
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						if attr.Val != "" {
							// there are three href cases
							// * same domain, full URLs
							// * same domain, relative path
							// * external domain, full URLs
							// let's find out what case this is
							ur, err := url.Parse(attr.Val)
							if err != nil {
								break
							}
							// remove fragments
							ur.Fragment = ""
							if ur.Host == loc.Host {
								// this is a same domain, full URL, we can use it
								result.LinkURLs[ur.String()]++
							} else if ur.Host == "" {
								// this is a same domain, relative path
								ur.Host = loc.Host
								ur.Scheme = loc.Scheme
								result.LinkURLs[ur.String()]++
							} else if c.crawlExternal {
								// these should all be external domains
								result.LinkURLs[ur.String()]++
							}
						}
						break
					}
				}
			case atom.Img, atom.Video, atom.Audio,
				atom.Script, atom.Style, atom.Link:
				// other elements can be used to extract asset URLs
				// TODO(geoah) Complete list of asset elements
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						result.AssetURLs[attr.Val]++
					}
				}
			}
			tokenType = doc.Next()
			continue
		}
		tokenType = doc.Next()
	}
	resp.Body.Close()
	return result, nil
}
