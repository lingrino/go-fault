package fault_test

import (
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

func TestHandlerError(t *testing.T) {
	t.Parallel()

	f := fault.New(fault.Options{
		Enabled:           true,
		Type:              "ERROR",
		Value:             500,
		PercentOfRequests: 1.0,
	})

	rr := sendRequest(t, f)

	if rr.Code != 500 {
		t.Errorf("wrong status code. expected: %v got: %v", 500, rr.Code)
	}

	if rr.Body.String() != "" {
		t.Errorf("wrong body. expected: %v got: %v", "", rr.Body.String())
	}

	f = fault.New(fault.Options{
		Enabled:           true,
		Type:              "ERROR",
		Value:             500,
		PercentOfRequests: 0.75,
	})

	var errorC, totalC float32

	for totalC <= 100000 {
		rr := sendRequest(t, f)
		if rr.Code == 500 {
			errorC++
		}
		totalC++
	}

	per := errorC / totalC

	if per > 0.76 || per < 0.74 {
		t.Errorf("wrong distribution. expected: 0.74 < per < 0.76, got: %v", per)
	}
}
