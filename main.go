package main

import (
	"fmt"
	"net/http"

	"github.com/geoah/go-crawl/crawler"
)

func main() {
	cl := &http.Client{}
	cr := crawler.New(cl, false)

	ws := &Website{
		crawler: cr,
		results: map[string]*crawler.Result{},
	}

	ws.Crawl("http://tomblomfield.com")
	for page, re := range ws.GetResults() {
		fmt.Println("\n=== ", page, " ===========")
		for _, url := range re.GetAssetURLs(true) {
			fmt.Println("Asset: ", url)
		}
		for _, url := range re.GetLinkURLs(true) {
			fmt.Println("Link: ", url)
		}
	}
}
