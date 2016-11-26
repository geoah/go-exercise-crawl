package main

import (
	"fmt"
	"net/http"
	"runtime"

	"time"

	"github.com/geoah/go-crawl/crawler"
)

func main() {
	cl := &http.Client{}
	cr := crawler.New(cl)

	nw := runtime.NumCPU()
	now := time.Now()

	// start crawling the website
	res, _ := cr.Crawl("http://tomblomfield.com", nw)

	// we can either get targets as soon as they have been processed
	for target := range res.StreamTargets() {
		fmt.Printf("\n=== %s ===========\n", target.GetURL().String())
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

	// or wait until we are done and get them all together
	targets := res.GetTargets()
	fmt.Printf("\n=== We fetched a total of %d pages using %d workers in %.2f seconds ===========",
		len(targets), nw, time.Since(now).Seconds())
}
