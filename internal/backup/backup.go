package backup

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/mailru/easyjson"
	"go.uber.org/multierr"

	"github.com/ASRafalsky/telemetry/internal/transport"
)

func DumpRepoToFile(path string, repo dataDumper, mode os.FileMode) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), addXPerm(mode)); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}

	defer func() {
		if errSync := f.Sync(); errSync != nil {
			err = multierr.Append(err, errSync)
		}
		if errClose := f.Close(); errClose != nil {
			err = multierr.Append(err, errClose)
		}
	}()

	zw := gzip.NewWriter(f)
	defer func() {
		if errClose := zw.Close(); errClose != nil {
			err = multierr.Append(err, errClose)
		}
	}()
	return dump(zw, repo)
}

func dump(w writer, repo dataDumper) error {
	if repo.Size() == 0 {
		return errors.New("repository is empty")
	}
	if err := repo.ForEach(context.Background(), func(k string, v []byte) error {
		v = append(v, '\n')
		_, err := w.Write(v)
		return err
	}); err != nil {
		return err
	}
	return nil
}

func RestoreRepoFromFile(path string, repo dataRestorer, remove bool) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			err = multierr.Append(err, errClose)
		}
	}()

	zr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := zr.Close(); errClose != nil {
			err = multierr.Append(err, errClose)
		}
	}()
	err = restore(zr, repo)

	if remove {
		if errRemove := os.Remove(path); errRemove != nil {
			err = multierr.Append(err, errRemove)
		}
	}

	return err
}

func restore(r reader, repo dataRestorer) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		buf := slices.Clone(scanner.Bytes())
		m := transport.Metrics{}
		if err := easyjson.Unmarshal(buf, &m); err != nil {
			return err
		}
		repo.Set(m.MType+m.ID, buf)
	}
	return scanner.Err()
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

type writer interface {
	Write(p []byte) (n int, err error)
}

type reader interface {
	Read(p []byte) (n int, err error)
}

type Decoder interface {
	Decode(e any) error
}
