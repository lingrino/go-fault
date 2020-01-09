package fault_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/github/fault"
)

// TestHandlerDisabled tests that the request proceeds normally when
// our middleware is disabled
func TestHandlerDisabled(t *testing.T) {
	t.Parallel()

	f := fault.New(fault.Options{
		Enabled: false,
	})

	rr := sendRequest(t, f)

	if rr.Code != testHandlerCode {
		t.Errorf("wrong status code. expected: %v got: %v", testHandlerCode, rr.Code)
	}

	if rr.Body.String() != testHandlerBody {
		t.Errorf("wrong body. expected: %v got: %v", testHandlerBody, rr.Body.String())
	}
}

// TestHandlerInvalidType tests that the request proceeds normally when
// we provide an invalid type
func TestHandlerInvalidType(t *testing.T) {
	t.Parallel()

	f := fault.New(fault.Options{
		Enabled: true,
		Type:    "INVALID",
	})

	rr := sendRequest(t, f)

	if rr.Code != testHandlerCode {
		t.Errorf("wrong status code. expected: %v got: %v", testHandlerCode, rr.Code)
	}

	if rr.Body.String() != testHandlerBody {
		t.Errorf("wrong body. expected: %v got: %v", testHandlerBody, rr.Body.String())
	}
}

func TestHelperLog(t *testing.T) {
	t.Parallel()

	// Test with a custom logger
	f := fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.Error,
		Value:             500,
		PercentOfRequests: 50,
		Debug:             true,
		Logger:            log.New(ioutil.Discard, "PREFIX", log.LstdFlags),
	})

	rr := sendRequest(t, f)

	if rr.Code != testHandlerCode {
		t.Errorf("wrong status code. expected: %v got: %v", testHandlerCode, rr.Code)
	}

	if rr.Body.String() != testHandlerBody {
		t.Errorf("wrong body. expected: %v got: %v", testHandlerBody, rr.Body.String())
	}

	// Test with the standard logger
	log.SetOutput(ioutil.Discard)
	f = fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.Error,
		Value:             500,
		PercentOfRequests: 50,
		Debug:             true,
	})

	rr = sendRequest(t, f)

	if rr.Code != testHandlerCode {
		t.Errorf("wrong status code. expected: %v got: %v", testHandlerCode, rr.Code)
	}

	if rr.Body.String() != testHandlerBody {
		t.Errorf("wrong body. expected: %v got: %v", testHandlerBody, rr.Body.String())
	}
}

// TestHandlerPercentDo indirectly tests the percentDo helper function by running an ERROR fault
// injection with different percents and validating that the faults occur at approximately those percents
//
// NOTE: Except for this test all other tests should use 0.0 or 1.0 for percentRequests
//       so that we have deterministic results and we don't test percentDo in multiple places
func TestHelperPercentDo(t *testing.T) {
	t.Parallel()

	// allowableRange is added/subtracted from percentExpected to get the allowed +/-
	// deviation from the expected percent. We're allowing a .5% deviation
	cases := []struct {
		percentRequests float32
		percentExpected float32
		allowableRange  float32
	}{
		{1.1, 0.0, 0},
		{1.0, 1.0, 0},
		{0.75, 0.75, 0.005},
		{0.3298, 0.3298, 0.005},
		{0.0001, 0.0001, 0.005},
		{0.0, 0.0, 0},
		{-0.1, 0.0, 0},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%g", tc.percentRequests), func(t *testing.T) {
			t.Parallel()

			f := fault.New(fault.Options{
				Enabled:           true,
				Type:              fault.Error,
				Value:             500,
				PercentOfRequests: tc.percentRequests,
			})

			var errorC, totalC float32

			for totalC <= 100000 {
				rr := sendRequest(t, f)
				if rr.Code == 500 {
					errorC++
				}
				totalC++
			}

			minP := tc.percentExpected - tc.allowableRange
			per := errorC / totalC
			maxP := tc.percentExpected + tc.allowableRange

			if per < minP || per > maxP {
				t.Errorf("wrong distribution. expected: %v < per < %v, got: %v", minP, maxP, per)
			}
		})
	}
}

// TestHandlerReject tests how we handle faults of the REJECT type. We only need to run one
// test with 0% chance and one with 100% for full coverage.
func TestHandlerReject(t *testing.T) {
	t.Parallel()

	cases := []struct {
		percentOfRequests float32
	}{
		{0.0},
		{1.0},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.percentOfRequests), func(t *testing.T) {

			f := fault.New(fault.Options{
				Enabled:           true,
				Type:              fault.Reject,
				PercentOfRequests: tc.percentOfRequests,
			})

			var rr *httptest.ResponseRecorder
			if tc.percentOfRequests == 1.0 {
				rr = sendRequestExpectPanic(t, f)
			} else {
				rr = sendRequest(t, f)
			}

			if rr != nil && tc.percentOfRequests == 1.0 {
				t.Errorf("expected: nil request got: %v", rr)
			}

			if rr != nil && rr.Code != testHandlerCode && tc.percentOfRequests == 0.0 {
				t.Errorf("wrong status code. expected: %v got: %v", testHandlerCode, rr.Code)
			}

			if rr != nil && rr.Body.String() != testHandlerBody && tc.percentOfRequests == 0.0 {
				t.Errorf("wrong body. expected: %v got: %v", testHandlerBody, rr.Body.String())
			}
		})
	}
}

// TestHandlerError tests how we handle faults of the ERROR type. We test with a bunch of
// valid and invalid error codes. With invalid codes we expect the handler to do nothing.
func TestHandlerError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sendCode   uint
		expectCode uint
		expectBody string
	}{
		{0, testHandlerCode, testHandlerBody},
		{1, testHandlerCode, testHandlerBody},
		{73, testHandlerCode, testHandlerBody},
		{100, 100, http.StatusText(100)},
		{199, testHandlerCode, testHandlerBody},
		{200, 200, http.StatusText(200)},
		{230, testHandlerCode, testHandlerBody},
		{404, 404, http.StatusText(404)},
		{500, 500, http.StatusText(500)},
		{501, 501, http.StatusText(501)},
		{600, testHandlerCode, testHandlerBody},
		{120000, testHandlerCode, testHandlerBody},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("Code %v", tc.sendCode), func(t *testing.T) {
			t.Parallel()

			f := fault.New(fault.Options{
				Enabled:           true,
				Type:              fault.Error,
				Value:             tc.sendCode,
				PercentOfRequests: 1.0,
			})

			rr := sendRequest(t, f)

			if rr.Code != int(tc.expectCode) {
				t.Errorf("wrong status code. expected: %v got: %v", tc.expectCode, rr.Code)
			}

			if strings.TrimSpace(rr.Body.String()) != tc.expectBody {
				t.Errorf("wrong body. expected: %v got: %v", tc.expectBody, rr.Body.String())
			}
		})
	}
}

// TestHandlerSlow tests how we handle faults of the SLOW type. Go time.Sleep()
// guarantees a sleep of AT LEAST the provided duration but potentially longer.
// In this package we strive to have no more than 5ms longer than requested.
// 5ms should be large enough to prevent flaky results on different machines.
func TestHandlerSlow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sendMs            uint
		expectMs          time.Duration
		allowableRange    time.Duration
		percentOfRequests float32
	}{
		{0, 0 * time.Millisecond, 1 * time.Millisecond, 0.0},
		{0, 0 * time.Millisecond, 1 * time.Millisecond, 1.0},
		{1, 1 * time.Millisecond, 3 * time.Millisecond, 1.0},
		{10, 10 * time.Millisecond, 5 * time.Millisecond, 1.0},
		{39, 39 * time.Millisecond, 5 * time.Millisecond, 1.0},
		{75, 75 * time.Millisecond, 5 * time.Millisecond, 1.0},
	}

	// First measure the time it takes to run with a 1ms wait, so that we
	// can substract "the speed of the system" from the "correct sleep time"
	f := fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.Slow,
		Value:             1,
		PercentOfRequests: 1.0,
	})

	t0 := time.Now()
	sendRequest(t, f)
	benchD := time.Since(t0) - 1*time.Millisecond

	for _, tc := range cases {
		tc := tc

		t.Run(fmt.Sprintf("%v", tc.sendMs), func(t *testing.T) {
			t.Parallel()

			f := fault.New(fault.Options{
				Enabled:           true,
				Type:              fault.Slow,
				Value:             tc.sendMs,
				PercentOfRequests: tc.percentOfRequests,
			})

			t0 := time.Now()
			rr := sendRequest(t, f)
			took := time.Since(t0)

			minD := time.Duration(tc.expectMs)
			maxD := time.Duration(tc.expectMs) + tc.allowableRange + benchD

			if took < minD || took > maxD {
				t.Errorf("wrong latency duration. expected: %v < duration < %v got: %v", minD, maxD, took)
			}

			if rr.Code != testHandlerCode {
				t.Errorf("wrong status code. expected: %v got: %v", testHandlerCode, rr.Code)
			}

			if rr.Body.String() != testHandlerBody {
				t.Errorf("wrong body. expected: %v got: %v", testHandlerBody, rr.Body.String())
			}
		})
	}
}

// TestHandlerChained tests how we handle chained faults, where we run one fault after
// another and we expect the second fault to ALWAYS activate if the first fault activated.
func TestHandlerChained(t *testing.T) {
	t.Parallel()

	// First test that the chained fault ALWAYS runs when the first fault is injected.
	t.Run("always", func(t *testing.T) {
		t.Parallel()

		f1 := fault.New(fault.Options{
			Enabled:           true,
			Type:              fault.Slow,
			Value:             10,
			PercentOfRequests: 1.0,
		})

		f2 := fault.New(fault.Options{
			Enabled:           true,
			Type:              fault.Error,
			Value:             500,
			PercentOfRequests: 0.0,
			Chained:           true,
		})

		t0 := time.Now()
		rr := sendRequest(t, f2, f1)
		took := time.Since(t0)

		minD := time.Duration(10 * time.Millisecond)
		maxD := time.Duration(13 * time.Millisecond)

		if took < minD || took > maxD {
			t.Errorf("wrong latency duration. expected: %v < duration < %v got: %v", minD, maxD, took)
		}

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("wrong status code. expected: %v got: %v", http.StatusInternalServerError, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != http.StatusText(500) {
			t.Errorf("wrong body. expected: %v got: %v", http.StatusText(500), rr.Body.String())
		}

	})

	// Next test that the chained fault NEVER runs when the first fault is not injected
	t.Run("never", func(t *testing.T) {
		t.Parallel()

		f1 := fault.New(fault.Options{
			Enabled:           true,
			Type:              fault.Slow,
			Value:             10,
			PercentOfRequests: 0.0,
		})

		f2 := fault.New(fault.Options{
			Enabled:           true,
			Type:              fault.Error,
			Value:             500,
			PercentOfRequests: 1.0,
			Chained:           true,
		})

		t0 := time.Now()
		rr := sendRequest(t, f2, f1)
		took := time.Since(t0)

		maxD := time.Duration(3 * time.Millisecond)

		if took > maxD {
			t.Errorf("wrong latency duration. expected: duration < %v got: %v", maxD, took)
		}

		if rr.Code != testHandlerCode {
			t.Errorf("wrong status code. expected: %v got: %v", testHandlerCode, rr.Code)
		}

		if rr.Body.String() != testHandlerBody {
			t.Errorf("wrong body. expected: %v got: %v", testHandlerBody, rr.Body.String())
		}

	})
}
