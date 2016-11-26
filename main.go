package main

import (
	"fmt"
	"net/http"

	"github.com/geoah/go-crawl/crawler"
)

func main() {
	cl := &http.Client{}
	cr := crawler.New(cl)

	res, _ := cr.Crawl("http://tomblomfield.com")
	for page, re := range res.GetResults() {
		fmt.Println("\n=== ", page, " ===========")
		for _, url := range re.GetAssetURLs(true) {
			fmt.Println("Asset: ", url)
		}
		for _, url := range re.GetLinkURLs(true) {
			fmt.Println("Link: ", url)
		}
	}
}
