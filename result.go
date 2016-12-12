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

		// TODO(geoah) Simplify flow

		// get the normalized url
		sURL := target.String()
		// check if we have already processed it
		if _, exists := r.urls[sURL]; exists {
			// if the target has an error it means that we already tried to
			// process it and failed
			if target.err != nil {
				// check if the error was retriable
				if ferr, ok := target.err.(FetcherError); ok && ferr.ShouldRetry() == false {
					// if the error is not retriable, mark the target as done,
					// push it to the results channel, and move on.
					target.done = true
					r.results <- target
					r.Done()
					continue
				}
			} else {
				// if we get a target that has already been processed
				// and there is no error, we shouldn't re-add to the queue.
				continue
			}
		} else {
			// if the target has never been processed, mark it as such
			r.urls[sURL] = true
			// bump the waitgroup
			r.Add(1)
		}

		// and finally push it to the queue to be processed
		r.jobs <- target

		// TODO(geoah) // HACK Need a better way to close channels
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
