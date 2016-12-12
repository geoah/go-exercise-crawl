package crawler

import "io"

// Parser attemps to parse the io.Reader from a Fetcher.
// Other resources that require futher parsing will be appended to the
// target's links map.
type Parser interface {
	Parse(target *Target, body io.Reader) error
}
