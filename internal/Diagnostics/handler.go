package diagnostics

import (
	"context"
	"net/http"
)

func ProfilePostHandler(ctx context.Context, path string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		go func() {
			_ = writeProfile(ctx, path+"heap", profileMem) // And I did it again.
		}()
		go func() {
			_ = writeProfile(ctx, path+"cpu_base.pprof", profileCPU) // And I did it again.
		}()
		res.WriteHeader(http.StatusOK)
	}
}

func MemProfilePostHandler(ctx context.Context, path string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		go func() {
			_ = writeProfile(ctx, path+"heap", profileMem) // And I did it again.
		}()
		res.WriteHeader(http.StatusOK)
	}
}

func CPUProfilePostHandler(ctx context.Context, path string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		go func() {
			_ = writeProfile(ctx, path+"cpu_base.pprof", profileCPU) // And I did it again.
		}()
		// I hope!))
		res.WriteHeader(http.StatusOK)
	}
}
