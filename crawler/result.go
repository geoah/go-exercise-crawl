package crawler

import "sync"

// Result is a complete crawl of a site
type Result struct {
	sync.WaitGroup
	queue   chan *Target
	jobs    chan *Target
	results chan *Target
	urls    map[string]bool
}

// Process is responsible of adding targets to the jobs channel
// to be processed by the workers.
// In case the workers fail to process the target, they will be
// re-added to the queue with an error where Process will check if
// it should retry or give up.
func (r *Result) Process() {
	// TODO(geoah) This is a hack to only wait and close channels after
	// at we have 1 item in the waitgroup.
	dc := false

	// now we can go through the queue as targets come in
	for target := range r.queue {
		// if a worker has successfully processed the target
		// we can mark is as Done and push it to the results.
		if target.done == true {
			r.results <- target
			r.Done()
			continue
		}

		// TODO(geoah) Refactor flow

		// else get the normalized url,
		sURL := target.String()
		// check if we have already processed this target
		if _, exists := r.urls[sURL]; exists {
			// and is there was an error check if we should retry
			if target.err != nil {
				if ferr, ok := target.err.(FetcherError); ok && ferr.ShouldRetry() {
					// and if so, re-add it to the jobs queue
				} else {
					// else just push it to targets and let them deal with the error
					target.done = true
					r.results <- target
					r.Done()
					continue
				}
			} else {
				// if the target has no error and we have already processed
				// this URL, we shouldn't re-add to the queue.
				continue
			}
		} else {
			// if not, mark it as processed
			r.urls[sURL] = true
			// bump the waitgroup
			r.Add(1)
		}

		// and finally push it to the queue to be processed
		r.jobs <- target

		// TODO(geoah) Need a better way to close channels
		if dc == false {
			dc = true
			go func(r *Result) {
				r.Wait()
				close(r.queue)
				close(r.jobs)
				close(r.results)
			}(r)
		}
	}
}
