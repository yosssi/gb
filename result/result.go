package result

import (
	"time"
)

// A Result represents a result of a HTTP request.
type Result struct {
	StartT time.Time
	EndT time.Time
	Error error
	HTTPStatusCode int
}

// Duration returns the result's duration.
func (r *Result) Duration() time.Duration {
	return r.EndT.Sub(r.StartT)
}

// Millisecond returns the result's duration in Milliseconds.
func (r *Result) Millisecond() int {
	return int(r.Duration() / time.Millisecond)
}


