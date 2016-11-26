# go-crawl

A very simple single domain crawler.

Given a URL, it will recursively go through all the pages under the original
hostname and will attempt to identify all static assets and links for each page.  

## To do

* [x] Merge website & crawler
* [x] Add worker pool instead of spawning everything in a go routine
* [ ] Move link parser under a Parser interface [low]
* [ ] Move HTTP downloader under a Getter interface [low]
* [ ] Add retry mechanism for failed GETs
* [ ] Add tests

## Known issues

These are issues that we can live with for now.

* We depend on the golang HTML parser to correct any invalid HTML.
* Only parses the response body for URLs.  
  Any URLs or assets that are inside of js/css files or generated by js code
  will not be detected. - golang phantomjs bindings maybe?
* Does not consolidate pages with same content but different query params.
* Does not check for URL schema, might cause same URL to be crawler under http & https.