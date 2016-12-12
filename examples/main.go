package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	crawler "github.com/geoah/go-crawl"
)

func main() {
	cl := &http.Client{}
	fe := crawler.NewFetcherHTTP(cl, 5)
	pa := crawler.NewParserHTML(false)
	cr := crawler.New(fe, pa)

	count := 0
	now := time.Now()
	nw := runtime.NumCPU()

	// start crawling the website
	results, err := cr.Crawl("http://tomblomfield.com", nw)
	if err != nil {
		fmt.Println(err)
	}

	// we can start getting targets as soon as they have been processed
	for target := range results {
		fmt.Printf("\n=== %s ===========\n", target.GetURL().String())
		count++
		if target.GetError() != nil {
			fmt.Printf("Error: Could not get page.\n")
			continue
		}
		for _, url := range target.GetAssetURLs(true) {
			fmt.Printf("Asset: %s\n", url)
		}
		for _, url := range target.GetLinkURLs(true) {
			fmt.Printf("Link: %s\n", url)
		}
	}

	fmt.Printf("\n=== We fetched a total of %d pages using %d workers in "+
		"%.2f seconds ===========",
		count, nw, time.Since(now).Seconds())
}
