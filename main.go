package main

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/geoah/go-crawl/crawler"
)

func main() {
	cl := &http.Client{}
	cr := crawler.New(cl)

	nw := runtime.NumCPU()

	res, _ := cr.Crawl("http://tomblomfield.com", nw)
	for page, target := range res.GetTargets() {
		fmt.Printf("\n=== %s ===========\n", page)
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
}
