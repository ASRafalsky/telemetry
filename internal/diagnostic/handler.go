// Package diagnostic contains useful things for profiling.
package diagnostic

import (
	"context"
	"net/http"
	"time"
)

// ProfilePostHandler starts profiling of app.
func ProfilePostHandler(ctx context.Context, path string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		_ = writeProfile(ctx, path+"heap-"+time.Now().Format(time.RFC3339), profileMem) // And I did it again.
		go func() {
			_ = writeProfile(ctx, path+"cpu-"+time.Now().Format(time.RFC3339), profileCPU) // And I did it again.
		}()
		res.WriteHeader(http.StatusOK)
	}
}
