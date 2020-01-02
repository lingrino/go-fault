package fault

import (
	"fmt"
	"net/http"
	"time"
)

const (
	// TypeDrop throws away the request with no response
	TypeDrop = "DROP"
	// TypeReject sends an empty reply back
	TypeReject = "REJECT"
	// TypeError injects http errors into the request
	TypeError = "ERROR"
	// TypeSlow injects latency into the request
	TypeSlow = "SLOW"
)

// Options is a struct for specifying all configuration
// to the fault.Fault middleware
type Options struct {
	// Set to true
	Enabled bool

	// Slow, Error, Drop, Reject
	Type string

	// Drop: n/a
	// Reject n/a
	// Error: http error to return
	// Slow: ms to wait
	Value int

	// How many requests to do this to
	PercentOfRequests float64
}

// Fault is middleware that does fault injection on incoming
// requests based on the configured options
type Fault struct {
	// Add options to determine what happens
	opt Options
}

// New returns a new Fault middleware with the supplied options
func New(o Options) *Fault {
	return &Fault{
		opt: o,
	}
}

// Handler implements http.HandlerFunc to use with net/http
func (f *Fault) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.opt.Enabled {
			stop, err := f.process(w, r)
			if err != nil {
				// Do not exit. Errors should not block
				// TODO - error types, handler here
				fmt.Println(err)
			}
			if stop {
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}

func (f *Fault) process(w http.ResponseWriter, r *http.Request) (bool, error) {
	var stop bool
	var err error

	switch f.opt.Type {
	case TypeDrop:
		stop, err = f.processReject(w, r)
	case TypeReject:
		stop, err = f.processReject(w, r)
	case TypeError:
		stop, err = f.processError(w, r)
	case TypeSlow:
		stop, err = f.processSlow(w, r)
	default:
		return stop, fmt.Errorf("failure type not specified")
	}

	return stop, err
}

func (f *Fault) processDrop(w http.ResponseWriter, r *http.Request) (bool, error) {
	var stop bool
	var err error

	if percentDo(f.opt.PercentOfRequests) {
		// I don't see a way to truly drop (never respond) an http request within the standard go
		// http stack. An option is http.Hijack() but it does not work on HTTP/2 connections. Dropped
		// requests in the context of fault injection are useful for testing timeouts, and we can
		// simulate that by sleeping for a long (5 minutes) period of time before eventuall returning
		// an empty response.
		time.Sleep(300 * time.Second)

		// This is a specialized and documented way of sending an interrupted
		// response to the client without printing the panic stack trace or erroring.
		// https://golang.org/pkg/net/http/#Handler
		panic(http.ErrAbortHandler)
	}

	return stop, err
}

func (f *Fault) processReject(w http.ResponseWriter, r *http.Request) (bool, error) {
	var stop bool
	var err error

	if percentDo(f.opt.PercentOfRequests) {
		// This is a specialized and documented way of sending an interrupted
		// response to the client without printing the panic stack trace or erroring.
		// https://golang.org/pkg/net/http/#Handler
		panic(http.ErrAbortHandler)
	}

	return stop, err
}

func (f *Fault) processError(w http.ResponseWriter, r *http.Request) (bool, error) {
	var stop bool
	var err error

	if percentDo(f.opt.PercentOfRequests) {
		// Do nothing if we don't have a valid status code
		if http.StatusText(f.opt.Value) == "" {
			return stop, err
		}

		w.WriteHeader(f.opt.Value)
		return true, nil
	}

	return stop, err
}

func (f *Fault) processSlow(w http.ResponseWriter, r *http.Request) (bool, error) {
	var stop bool
	var err error

	if percentDo(f.opt.PercentOfRequests) {
		time.Sleep(time.Duration(f.opt.Value) * time.Millisecond)
	}

	return stop, err
}
