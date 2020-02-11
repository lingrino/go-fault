package fault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUpdateRequestContextValue tests updateRequestContextValue in isolation.
func TestUpdateRequestContextValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		giveReq     *http.Request
		giveReqCtx  context.Context
		giveCtxList ContextValue
		wantCtxList ContextValue
		wantNilReq  bool
	}{
		{
			name:        "nil",
			giveReq:     nil,
			giveReqCtx:  context.Background(),
			giveCtxList: ContextValue{ContextValueInjected},
			wantCtxList: nil,
			wantNilReq:  true,
		},
		{
			name:        "one value",
			giveReq:     &http.Request{},
			giveReqCtx:  context.Background(),
			giveCtxList: ContextValue{ContextValueInjected},
			wantCtxList: ContextValue{ContextValueInjected},
		},
		{
			name:        "two values",
			giveReq:     &http.Request{},
			giveReqCtx:  context.Background(),
			giveCtxList: ContextValue{ContextValueInjected, ContextValueSlowInjector},
			wantCtxList: ContextValue{ContextValueInjected, ContextValueSlowInjector},
		},
		{
			name:        "with existing values",
			giveReq:     &http.Request{},
			giveReqCtx:  context.WithValue(context.Background(), ContextKey, ContextValue{ContextValueError}),
			giveCtxList: ContextValue{ContextValueInjected, ContextValueSlowInjector},
			wantCtxList: ContextValue{ContextValueError, ContextValueInjected, ContextValueSlowInjector},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var req *http.Request
			if !tt.wantNilReq {
				req = tt.giveReq.WithContext(tt.giveReqCtx)
			}

			for _, cs := range tt.giveCtxList {
				req = updateRequestContextValue(req, cs)
			}

			if tt.wantNilReq {
				assert.Nil(t, req)
			} else {
				gotCtx := req.Context().Value(ContextKey)
				assert.Equal(t, tt.wantCtxList, gotCtx)
			}
		})
	}
}

// TestUpdateRequestContextValue_WithHandler tests updateRequestContextValue within a handler.
func TestUpdateRequestContextValue_WithHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		giveHandler func(http.Handler) http.Handler
		wantCtxList ContextValue
	}{
		{
			name: "inline update",
			giveHandler: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, updateRequestContextValue(r, ContextValueInjected))
				})
			},
			wantCtxList: ContextValue{ContextValueInjected},
		},
		{
			name: "assign before",
			giveHandler: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r = updateRequestContextValue(r, ContextValueInjected)
					next.ServeHTTP(w, r)
				})
			},
			wantCtxList: ContextValue{ContextValueInjected},
		},
		{
			name: "multiple",
			giveHandler: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r = updateRequestContextValue(r, ContextValueInjected)
					r = updateRequestContextValue(r, ContextValueSkipped)
					next.ServeHTTP(w, r)
				})
			},
			wantCtxList: ContextValue{ContextValueInjected, ContextValueSkipped},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctxVerifier := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.wantCtxList, r.Context().Value(ContextKey))
			})

			chain := tt.giveHandler(ctxVerifier)

			chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
		})
	}
}
