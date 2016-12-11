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

// Crawler allows us to recursively crawl something given
// a fetcher and a parser.
// The most common use case would be to use an HTTP Fetcher
// and an HTML Parser to crawl a website.
// Beside that we can create new Fetchers/Parsers to crawl
// twitter feeds, slack channels, custom apis, or just write
// parsers for specific usecases.
type Crawler struct {
	fetcher Fetcher
	parser  Parser
}

// New crawler from a fether and a parser
func New(fetcher Fetcher, parser Parser) *Crawler {
	return &Crawler{
		fetcher: fetcher,
		parser:  parser,
	}
}

// Crawl starts starts a given number of worker go routines, and adds the
// targetURL to the queue to be fetched and parsed.
// Once the targetURL has been parsed, if there are additional links that
// need fetching, they will be added to the queue as well.
// As soon as a link has been parsed, it will be pushed to the `result.results`
// channel. After we are done will all possible links, the channel will close.
func (c *Crawler) Crawl(targetURL string, workers int) (*Result, error) {
	// initialize our result
	result := &Result{
		urls:    map[string]bool{},
		queue:   make(chan *Target, 100),
		jobs:    make(chan *Target, 100),
		results: make(chan *Target, 100),
	}

	// create a target for our entrypoint
	tgt, err := NewTarget(targetURL)
	if err != nil {
		return nil, err
	}

	// start our queue/queue processing
	go result.Process()

	// add our entrypoint url to our queue
	result.queue <- tgt

	// make sure we have some workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}

	// start our workers
	for w := 1; w <= workers; w++ {
		go c.worker(result.queue, result.jobs)
	}

	return result, nil
}

func (c *Crawler) worker(queue, jobs chan *Target) {
	// go through our target queue
	for target := range jobs {
		// make sure that we remove it from the waitgroup when all is done
		// try to fetch the content of target
		body, err := c.fetcher.Fetch(target)
		// and if something didn't go as expected
		if err != nil {
			target.err = err
			queue <- target
			continue
		}
		// if everything was ok, we can try to parse the content
		// we currently don't really handle the parser failing
		// TODO(geoah) Maybe re-try parsing if it fails?
		if err = c.parser.Parse(target, body); err == nil {
			// if there are links add them to the queue
			for _, nURL := range target.GetLinkURLs(true) {
				ntgt, err := NewTarget(nURL)
				if err != nil {
					ntgt.err = err
				}
				queue <- ntgt
				continue
			}
		}
		// mark target as done
		target.done = true
		queue <- target
	}
}
