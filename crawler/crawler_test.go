package crawler

import (
	"net/http"
	"testing"

	"fmt"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
)

type TestEndpoint struct {
	code int
	body string
}

type CrawlerTestSuite struct {
	suite.Suite
	client     *http.Client
	crawler    *Crawler
	maxRetries int
	endpoints  map[string]TestEndpoint
}

func TestCrawlerTestSuite(t *testing.T) {
	suite.Run(t, new(CrawlerTestSuite))
}

func (suite *CrawlerTestSuite) SetupTest() {
	// init our client and crawler
	suite.client = &http.Client{}
	suite.maxRetries = 5
	suite.crawler = &Crawler{
		client:     suite.client,
		maxRetries: suite.maxRetries,
	}
	suite.endpoints = map[string]TestEndpoint{}

	// create some fake endpoints
	suite.endpoints["http://test/"] = TestEndpoint{200, `<a href="http://test/a">a</a><a href="/b">b</a>`}
	suite.endpoints["http://test/a"] = TestEndpoint{200, `<a href="http://test/c">c</a>`}
	suite.endpoints["http://test/b"] = TestEndpoint{200, `<a href="http://test/c">c</a>`}
	suite.endpoints["http://test/c"] = TestEndpoint{200, `<a href="http://test/404">404</a>`}
	suite.endpoints["http://test/loop"] = TestEndpoint{200, `<a href="http://test/loop">loop</a>`}
	suite.endpoints["http://test/204"] = TestEndpoint{404, ``}
	suite.endpoints["http://test/400"] = TestEndpoint{400, ``}
	suite.endpoints["http://test/401"] = TestEndpoint{401, ``}
	suite.endpoints["http://test/404"] = TestEndpoint{404, ``}
	suite.endpoints["http://test/429"] = TestEndpoint{429, ``}
	suite.endpoints["http://test/500"] = TestEndpoint{500, ``}

	// activate our mock
	httpmock.Activate()
	// and finally register our endpoints
	for ep, te := range suite.endpoints {
		httpmock.RegisterResponder("GET", ep, httpmock.NewStringResponder(te.code, te.body))
	}
}

func (suite *CrawlerTestSuite) TestCrawlInvalidURL() {
	// crawl an invalid endpoint
	res, err := suite.crawler.Crawl("not a valid url", 1)
	suite.Nil(res)
	suite.NotNil(err)
	suite.Equal(ErrInvalidURL, err)
}

func (suite *CrawlerTestSuite) TestCrawlRecursive() {
	// crawl a valid endpoints that leads to 2 other links
	res, err := suite.crawler.Crawl("http://test/", 1)
	suite.NotNil(res)
	suite.Nil(err)

	// we should have 5 targets (/, /a, /b, /c, /404 with /404 having errored)
	tgs := res.GetTargets()
	suite.Len(tgs, 5)

	suite.Len(tgs["http://test/"].GetLinkURLs(false), 2)
	suite.Nil(tgs["http://test/"].GetError())

	suite.Len(tgs["http://test/a"].GetLinkURLs(false), 1)
	suite.Nil(tgs["http://test/a"].GetError())

	suite.Len(tgs["http://test/b"].GetLinkURLs(false), 1)
	suite.Nil(tgs["http://test/b"].GetError())

	suite.Len(tgs["http://test/c"].GetLinkURLs(false), 1)
	suite.Nil(tgs["http://test/c"].GetError())

	suite.Len(tgs["http://test/404"].GetLinkURLs(false), 0)
	suite.NotNil(tgs["http://test/404"].GetError())
}

func (suite *CrawlerTestSuite) TestCrawlLoop() {
	// crawl a valid endpoints that leads to itself
	res, err := suite.crawler.Crawl("http://test/loop", 1)
	suite.NotNil(res)
	suite.Nil(err)

	// we should have 1 targets (/loop)
	tgs := res.GetTargets()
	suite.Len(tgs, 1)

	suite.Len(tgs["http://test/loop"].GetLinkURLs(false), 1)
	suite.Nil(tgs["http://test/loop"].GetError())
}

func (suite *CrawlerTestSuite) TestCrawlErrors() {
	// there are two types of failures
	// the ones we retry until maxRetries (429, 500)
	retryCodes := []int{429, 500}
	// the ones we instantly fail (400, 401, 404)
	failCodes := []int{400, 401, 404}

	// go through the codes where we retry
	for _, code := range retryCodes {
		ep := fmt.Sprintf("http://test/%d", code)

		// crawl a valid endpoints that retuns an error code
		res, err := suite.crawler.Crawl(ep, 1)
		suite.NotNil(res)
		suite.Nil(err)

		// we should have 1 target
		tgs := res.GetTargets()
		suite.Len(tgs, 1)

		// and they should error
		suite.Len(tgs[ep].GetLinkURLs(false), 0, tgs[ep].url.String())
		suite.Equal(suite.maxRetries, tgs[ep].tries, tgs[ep].url.String())
		suite.NotNil(tgs[ep].GetError(), tgs[ep].url.String())
	}

	// go through the codes we should fail
	for _, code := range failCodes {
		ep := fmt.Sprintf("http://test/%d", code)

		// crawl a valid endpoints that retuns an error code
		res, err := suite.crawler.Crawl(ep, 1)
		suite.NotNil(res)
		suite.Nil(err)

		// we should have 1 target
		tgs := res.GetTargets()
		suite.Len(tgs, 1)

		// and they should error
		suite.Len(tgs[ep].GetLinkURLs(false), 0, tgs[ep].url.String())
		suite.Equal(suite.maxRetries+1, tgs[ep].tries, tgs[ep].url.String())
		suite.NotNil(tgs[ep].GetError(), tgs[ep].url.String())
	}
}
