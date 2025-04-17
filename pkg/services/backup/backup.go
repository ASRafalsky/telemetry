package backup

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/multierr"
)

// Looks stupid enough store repository name for each entry, but...)))
type entry struct {
	Name  string
	Key   string
	Value []byte
}

func DumpRepoToFile(path string, repos map[string]Repository, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), addXPerm(mode)); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	return dump(gob.NewEncoder(f), repos)
}

func dump(enc Encoder, repos map[string]Repository) error {
	if len(repos) == 0 {
		return errors.New("no repository found")
	}

	for name, repo := range repos {
		if err := repo.ForEach(context.Background(), func(k string, v []byte) error {
			return enc.Encode(entry{
				Name:  name,
				Key:   k,
				Value: v,
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

func RestoreRepoFromFile(path string, repos map[string]Repository, remove bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	errRes := restore(gob.NewDecoder(f), repos)

	if remove {
		if err := os.Remove(path); err != nil {
			errRes = multierr.Append(errRes, err)
		}
	}

	return errRes
}

func restore(dec Decoder, repos map[string]Repository) error {
	var errRes error
	for {
		var repoEntry entry
		if err := dec.Decode(&repoEntry); err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		if _, ok := repos[repoEntry.Name]; ok {
			repos[repoEntry.Name].Set(repoEntry.Key, repoEntry.Value)
			continue
		}
		errRes = multierr.Append(errRes, fmt.Errorf("%s not found", repoEntry.Name))
	}
	return errRes
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

type Repository interface {
	Set(k string, v []byte)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
}

type Encoder interface {
	Encode(v any) error
}

type Decoder interface {
	Decode(e any) error
}
