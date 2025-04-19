package backup

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/mailru/easyjson"
	"go.uber.org/multierr"

	"github.com/ASRafalsky/telemetry/internal/transport"
)

func DumpRepoToFile(path string, repo repository, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), addXPerm(mode)); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Sync()
	defer f.Close()

	zw := gzip.NewWriter(f)
	defer zw.Close()
	return dump(zw, repo)
}

func dump(w writer, repo repository) error {
	if repo.Size() == 0 {
		return errors.New("repository is empty")
	}
	if err := repo.ForEach(context.Background(), func(k string, v []byte) error {
		_, err := w.Write(v)
		return err
	}); err != nil {
		return err
	}
	return nil
}

func RestoreRepoFromFile(path string, repo repository, remove bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	zr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer zr.Close()
	errRes := restore(zr, repo)

	if remove {
		if err := os.Remove(path); err != nil {
			errRes = multierr.Append(errRes, err)
		}
	}

	return errRes
}

func restore(r reader, repo repository) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	metrics, err := transport.DeserializeMetrics(buf)
	if err != nil {
		return err
	}
	for _, m := range metrics {
		buf, err := easyjson.Marshal(m)
		if err != nil {
			return err
		}
		repo.Set(m.MType+m.ID, buf)
	}
	return nil
}

func addXPerm(mode os.FileMode) os.FileMode {
	const (
		permGroups = 3
		permBitSzPerGroup
	)
	rPerm := uint32(4) // r--
	xPerm := uint32(1) // --x
	for i := range permGroups {
		if (mode & os.FileMode(rPerm<<(i*permBitSzPerGroup))) != 0 {
			mode |= os.FileMode(xPerm << (i * permBitSzPerGroup))
		}
	}

	return mode.Perm()
}

type repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	Delete(k string)
}

type writer interface {
	Write(p []byte) (n int, err error)
}

type reader interface {
	Read(p []byte) (n int, err error)
}

type Decoder interface {
	Decode(e any) error
}
