package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	// ErrInvalidURL trying to crawl an invalid URL
	ErrInvalidURL = errors.New("Invalid URL")
)

// Crawler is basically an HTTP scraper
type Crawler struct {
	client  *http.Client
	results map[string]*Result
	queue   chan string
}

// New creates a new Crawler from an HTTP client.
func New(client *http.Client) *Crawler {
	return &Crawler{
		client: client,
	}
}

func (c *Crawler) worker(id int, results *Results) {
	for target := range results.queue {
		fmt.Println("worker", id, "processing url", target)
		go func(target string, results *Results) {
			result, _ := c.process(target)
			results.Lock()
			results.results[target] = result
			results.Unlock()
			for _, ntarget := range result.GetLinkURLs(true) {
				results.Enqueue(ntarget)
			}
			results.Done()
		}(target, results)
	}
}

// Crawl starts a recursive crawl of a site
func (c *Crawler) Crawl(target string) (*Results, error) {
	// initialize our results
	results := &Results{
		results: map[string]*Result{},
		queue:   make(chan string, 100),
	}

	// TODO(geoah) Instead of Wait()ing in GetResults() maybe wait here?
	// make sure we wait until everything is done
	// defer results.Wait()

	// enqueue our entrypint url
	results.Enqueue(target)

	// start our workers
	nw := runtime.NumCPU()
	for w := 1; w <= nw; w++ {
		go c.worker(w, results)
	}

	return results, nil
}

// process -
func (c *Crawler) process(target string) (*Result, error) {
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
							}
							//  else if c.crawlExternal {
							// 	// these should all be external domains
							// 	result.LinkURLs[ur.String()]++
							// }
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
