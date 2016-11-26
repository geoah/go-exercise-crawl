package crawler

import (
	"errors"
	"runtime"
)

var (
	// ErrInvalidURL trying to crawl an invalid URL
	ErrInvalidURL = errors.New("Invalid URL")
	// ErrFetchingPage trying to fetch a page
	ErrFetchingPage = errors.New("Error fetching page")
)

// Crawler is basically an HTTP scraper
type Crawler struct {
	fetcher    Fetcher
	parser     Parser
	maxRetries int
}

// New creates a new Crawler from an HTTP client.
func New(fetcher Fetcher, parser Parser) *Crawler {
	return &Crawler{
		fetcher: fetcher,
		parser:  parser,
	}
}

// Crawl starts a recursive crawl of a site with a number of workers
func (c *Crawler) Crawl(targetURL string, workers int) (*Result, error) {
	// initialize our result
	result := &Result{
		urls:    map[string]bool{},
		queue:   make(chan *Target, 100),
		targets: make(chan *Target, 100),
	}

	// enqueue our entrypoint url
	if err := result.Enqueue(targetURL); err != nil {
		return nil, err
	}

	// make sure we have some workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}

	// start our workers
	for w := 1; w <= workers; w++ {
		go c.worker(result)
	}

	// when all is said and done, close everything up
	go func(result *Result) {
		result.Wait()
		close(result.queue)
		close(result.targets)
	}(result)

	return result, nil
}

func (c *Crawler) worker(result *Result) {
	for target := range result.queue {
		go func(target *Target, result *Result) {
			defer result.Done()
			// fetch body of target
			body, err := c.fetcher.Fetch(target)
			if err != nil {
				// check if we should retry
				if ferr, ok := err.(FetcherError); ok && ferr.ShouldRetry() {
					// and if so, re-add it to the end of the queue
					result.Add(1)
					result.queue <- target
					return
				}
				// else just push it to targets
				result.targets <- target
				return
			}
			// and parse it
			if err = c.parser.Parse(target, body); err == nil {
				// if there are links add them to the queue
				for _, ntarget := range target.GetLinkURLs(true) {
					result.Enqueue(ntarget)
				}
			}
			result.targets <- target
		}(target, result)
	}
}
