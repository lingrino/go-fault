package fault

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
