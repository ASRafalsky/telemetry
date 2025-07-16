package diagnostics

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

func writeProfile(ctx context.Context, path string, fn func(ctx context.Context, w io.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return err
	}

	var b bytes.Buffer
	if err := fn(ctx, &b); err != nil {
		return err
	}

	return os.WriteFile(path, b.Bytes(), 0o644)
}

func profileCPU(ctx context.Context, w io.Writer) error {
	if err := pprof.StartCPUProfile(w); err != nil {
		return fmt.Errorf("failed to start CPU profile; %w", err)
	}
	defer pprof.StopCPUProfile()

	const profPeriod = 30 * time.Second
	select {
	case <-time.After(profPeriod):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func profileMem(_ context.Context, w io.Writer) error {
	return pprof.WriteHeapProfile(w)
}
