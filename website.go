package main

import (
	"sync"

	"github.com/geoah/go-crawl/crawler"
)

// Website holds all the pages of a website
type Website struct {
	sync.WaitGroup
	sync.RWMutex
	crawler *crawler.Crawler
	results map[string]*crawler.Result
}

// Crawl -
func (ws *Website) Crawl(target string) {
	ws.Lock()
	if _, exists := ws.results[target]; exists {
		ws.Unlock()
		return
	}
	ws.results[target] = nil
	ws.Unlock()

	ws.Add(1)
	go func() {
		res, _ := ws.crawler.Crawl(target)
		ws.results[target] = res
		for _, url := range res.GetLinkURLs(true) {
			ws.Crawl(url)
		}
		ws.Done()
	}()
}

// GetResults -
func (ws *Website) GetResults() map[string]*crawler.Result {
	ws.Wait()
	return ws.results
}
