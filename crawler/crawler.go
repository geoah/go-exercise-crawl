package crawler

import (
	"errors"
	"net/http"
	"net/url"
	"runtime"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	// ErrInvalidURL trying to crawl an invalid URL
	ErrInvalidURL = errors.New("Invalid URL")
	// ErrFetchingPage trying to fetch a page
	ErrFetchingPage = errors.New("Error fetching page")
)

// Crawler is basically an HTTP scraper
type Crawler struct {
	client     *http.Client
	maxRetries int
}

// New creates a new Crawler from an HTTP client.
func New(client *http.Client) *Crawler {
	return &Crawler{
		client:     client,
		maxRetries: 5,
	}
}

// Crawl starts a recursive crawl of a site with a number of workers
func (c *Crawler) Crawl(targetURL string, workers int) (*Results, error) {
	// initialize our results
	results := &Results{
		results: map[string]*Target{},
		queue:   make(chan *Target, 100),
	}

	// TODO(geoah) Instead of Wait()ing in results.GetResults() maybe wait here?
	// make sure we wait until everything is done
	// defer results.Wait()

	// enqueue our entrypoint url
	if err := results.Enqueue(targetURL); err != nil {
		return nil, err
	}

	// make sure we have some workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}

	// start our workers
	for w := 1; w <= workers; w++ {
		go c.worker(w, results)
	}

	return results, nil
}

func (c *Crawler) worker(id int, results *Results) {
	for target := range results.queue {
		go func(target *Target, results *Results) {
			c.process(target)
			// TODO(geoah) It might be better to re-add the targets to the
			// queue instead of instantly retrying
			// if the url could not be processes and we can still retry
			for target.err != nil && target.tries < c.maxRetries {
				c.process(target)
			}
			// if there is no error we can go through the found links and
			// queue them for further processing
			if target.err == nil {
				for _, ntarget := range target.GetLinkURLs(true) {
					results.Enqueue(ntarget)
				}
			}
			results.Done()
		}(target, results)
	}
}

// process will try to get a target URL, parse it's body, and find all links and assets
// contained in the page. If something goes wrong it will populate the target's error.
func (c *Crawler) process(target *Target) {
	// bump tries and remove error
	target.tries++
	target.err = nil

	// construct and make our HTTP request
	req, err := http.NewRequest("GET", target.url.String(), nil)
	if err != nil {
		target.err = err
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		target.err = err
		return
	}
	// make sure to close body when all is done
	defer resp.Body.Close()

	// TODO(geoah) Handle other HTTP codes a bit more generically; eg 2xx, 4xx, 5xx
	// TODO(geoah) Handling 429 (too-many-requests) errors specially would also be nice
	// figure out if we need to parse the page or return an error
	switch resp.StatusCode {
	case http.StatusNoContent:
		// if there is no content, just return
		return
	case http.StatusNotFound,
		http.StatusUnauthorized,
		http.StatusBadRequest:
		// don't retry and just return an error
		// TODO(geoah) Maybe return a new ErrPageNotExist error
		target.tries = c.maxRetries + 1
		target.err = ErrFetchingPage
		return
	case http.StatusInternalServerError,
		http.StatusTooManyRequests:
		// return error but allow retry
		target.err = ErrFetchingPage
		return
	}

	// find out our location from headers as we might have been redirected
	loc, err := resp.Location()
	if err != nil {
		// if we can't we can always fall back to the original target
		loc = target.url
	}

	// since we managed to get this far, clean the error
	target.err = nil

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
								target.LinkURLs[ur.String()]++
							} else if ur.Host == "" {
								// this is a same domain, relative path
								ur.Host = loc.Host
								ur.Scheme = loc.Scheme
								target.LinkURLs[ur.String()]++
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
						target.AssetURLs[attr.Val]++
					}
				}
			}
			tokenType = doc.Next()
			continue
		}
		tokenType = doc.Next()
	}
	resp.Body.Close()
}
