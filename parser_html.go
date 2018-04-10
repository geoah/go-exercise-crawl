package crawler

import (
	"io"
	"net/url"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ParserHTML is a HTML document parser
type ParserHTML struct {
	followExternalLinks bool
}

// NewParserHTML creates a new ParserHTML
func NewParserHTML(followExternalLinks bool) *ParserHTML {
	return &ParserHTML{followExternalLinks}
}

// Parse will go through the document elements and will attempt to find all
// same-domain links that can be further crawled as well as all assets the
// target uses.
func (p *ParserHTML) Parse(target *Target, body io.Reader) error {
	loc := target.GetURL()
	doc := html.NewTokenizer(body)
	tokenType := doc.Next()
	for tokenType != html.ErrorToken {
		token := doc.Token()
		// we only care about starting elements
		if tokenType == html.StartTagToken {
			switch token.DataAtom {
			case atom.A:
				// we assume that <a> elements will provide us with all links
				// go through the attributes and find the HREF one
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						if attr.Val != "" {
							// there are three href cases
							// * same domain, full URLs
							// * same domain, relative path
							// * external domain, full URLs
							// let's find out what case this is
							ur, err := url.Parse(attr.Val)
							if err != nil {
								break
							}
							// remove fragments
							ur.Fragment = ""
							if p.followExternalLinks || ur.Host == loc.Host {
								// this is a same domain, full URL, we can use it
								target.linkURLs[ur.String()]++
							} else if ur.Host == "" {
								// this is a same domain, relative path
								ur.Host = loc.Host
								ur.Scheme = loc.Scheme
								target.linkURLs[ur.String()]++
							}
						}
						break
					}
				}
			case atom.Img, atom.Video, atom.Audio,
				atom.Script, atom.Style, atom.Link:
				// other elements can be used to extract asset URLs
				// TODO(geoah) Complete list of asset elements and attributes
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						target.assetURLs[attr.Val]++
					}
					if attr.Key == "href" {
						target.assetURLs[attr.Val]++
					}
				}
			}
			tokenType = doc.Next()
			continue
		}
		tokenType = doc.Next()
	}
	return nil
}
