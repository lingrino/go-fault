package fault_test

import (
	"fmt"
	"testing"

	"github.com/github/fault"
)

func TestHandlerDisabled(t *testing.T) {
	t.Parallel()

	fault := fault.New(fault.Options{
		Enabled: false,
	})

	rr := sendRequest(t, fault)

	if rr.Code != testHandlerCode {
		t.Errorf("wrong status code. expected: %v got: %v", testHandlerCode, rr.Code)
	}

	if rr.Body.String() != testHandlerBody {
		t.Errorf("wrong body. expected: %v got: %v", testHandlerBody, rr.Body.String())
	}
}

// TestHandlerPercentDo indirectly tests the percentDo helper function
// by running an ERROR fault injection with different percents and validating
// that the faults occur at approximately those percents
func TestHelperPercentDo(t *testing.T) {
	t.Parallel()

	cases := []struct {
		percentRequests float64
		percentExpected float64
		allowableRange  float64
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
				Type:              fault.TypeError,
				Value:             500,
				PercentOfRequests: tc.percentRequests,
			})

			var errorC, totalC float64

			for totalC <= 100000 {
				rr := sendRequest(t, f)
				if rr.Code == 500 {
					errorC++
				}
				totalC++
			}

			per := errorC / totalC

			if per > tc.percentExpected+tc.allowableRange || per < tc.percentExpected-tc.allowableRange {
				t.Errorf("wrong distribution. expected: %v < per < %v, got: %v", tc.percentExpected-tc.allowableRange, tc.percentExpected+tc.allowableRange, per)
			}
		})
	}
}

func TestHandlerError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sendCode   int
		expectCode int
		expectBody string
	}{
		{0, testHandlerCode, testHandlerBody},
		{1, testHandlerCode, testHandlerBody},
		{73, testHandlerCode, testHandlerBody},
		{100, 100, ""},
		{199, testHandlerCode, testHandlerBody},
		{200, 200, ""},
		{230, testHandlerCode, testHandlerBody},
		{404, 404, ""},
		{500, 500, ""},
		{501, 501, ""},
		{600, testHandlerCode, testHandlerBody},
		{120000, testHandlerCode, testHandlerBody},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("Code %v", tc.sendCode), func(t *testing.T) {
			t.Parallel()

			f := fault.New(fault.Options{
				Enabled:           true,
				Type:              fault.TypeError,
				Value:             tc.sendCode,
				PercentOfRequests: 1.0,
			})

			rr := sendRequest(t, f)

			if rr.Code != tc.expectCode {
				t.Errorf("wrong status code. expected: %v got: %v", tc.expectCode, rr.Code)
			}

			if rr.Body.String() != tc.expectBody {
				t.Errorf("wrong body. expected: %v got: %v", tc.expectBody, rr.Body.String())
			}
		})
	}
}

func TestHandlerReject(t *testing.T) {
	t.Parallel()

	f := fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.TypeReject,
		PercentOfRequests: 1.0,
	})

	rr := sendRequest(t, f)

	fmt.Println(rr)
}
