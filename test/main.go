package main

import (
	"net/http"

	"github.com/github/fault"
)

const (
	testHandlerCode        = http.StatusOK
	testHandlerContentType = "application/json"
	testHandlerBody        = `{"status": "OK"}`
)

var mainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(testHandlerCode)
	w.Header().Set("Content-Type", testHandlerContentType)
	w.Write([]byte(testHandlerBody))
})

func main() {

	rejectFault := fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.TypeReject,
		PercentOfRequests: 0.25,
	})

	errorFault := fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.TypeError,
		Value:             500,
		PercentOfRequests: 0.25,
	})

	slowFault := fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.TypeSlow,
		Value:             2000, // 2 seconds
		PercentOfRequests: 0.25,
	})

	handlerChain := slowFault.Handler(rejectFault.Handler(errorFault.Handler(mainHandler)))

	http.ListenAndServe("0.0.0.0:3000", handlerChain)
}
