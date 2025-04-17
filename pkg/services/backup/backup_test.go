package backup

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/storage"
)

func TestBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "/test/backup")

	repos := map[string]Repository{
		"first":  storage.New[string, []byte](),
		"second": storage.New[string, []byte](),
	}

	for i := range 100 {
		idx := strconv.Itoa(i)
		repos["first"].Set("first-"+idx, []byte("first-"+idx))
		repos["second"].Set("second-"+idx, []byte("second-"+idx))
	}

	// Add data to dump.
	require.NoError(t, DumpRepoToFile(path, repos, 0o644))

	// Check the dump file.
	stat, err := os.Stat(path)
	require.NoError(t, err)
	require.NotZero(t, stat.Size())

	restoredRepos := map[string]Repository{
		"first":  storage.New[string, []byte](),
		"second": storage.New[string, []byte](),
	}

	// Try to restore from the dump file.
	err = RestoreRepoFromFile(path, restoredRepos, false)
	require.NoError(t, err)
	require.NotZero(t, restoredRepos["first"].Size())
	require.NotZero(t, restoredRepos["second"].Size())
	require.Equal(t, repos["first"].Size(), restoredRepos["first"].Size())
	require.Equal(t, repos["second"].Size(), restoredRepos["second"].Size())
	require.Equal(t, repos["first"], restoredRepos["first"])
	require.Equal(t, repos["second"], restoredRepos["second"])

	restoredRepos2 := map[string]Repository{
		"first":  storage.New[string, []byte](),
		"second": storage.New[string, []byte](),
	}

	err = RestoreRepoFromFile(path, restoredRepos2, true)
	require.NoError(t, err)
	require.NotZero(t, restoredRepos2["first"].Size())
	require.NotZero(t, restoredRepos2["second"].Size())
	// Yes, you are doing everything right.)))
	_, err = os.Stat(path)
	require.Error(t, err, os.ErrNotExist)
}

func BenchmarkJSONDump(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	repos := map[string]Repository{
		"first":  storage.New[string, []byte](),
		"second": storage.New[string, []byte](),
	}
	restoredRepos := map[string]Repository{
		"first":  storage.New[string, []byte](),
		"second": storage.New[string, []byte](),
	}

	for i := range 1000 {
		idx := strconv.Itoa(i)
		repos["first"].Set("first-"+idx, []byte("first-"+idx))
		repos["second"].Set("second-"+idx, []byte("second-"+idx))
	}

	for i := 0; i < b.N; i++ {
		err1 := dump(json.NewEncoder(buf), repos)
		err2 := restore(json.NewDecoder(buf), restoredRepos)
		b.StopTimer()
		if err1 != nil {
			b.Fatal(err1)
		}
		if err2 != nil {
			b.Fatal(err2)
		}
		buf.Reset()
		b.StartTimer()
	}
}

func BenchmarkGOBDump(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	repos := map[string]Repository{
		"first":  storage.New[string, []byte](),
		"second": storage.New[string, []byte](),
	}
	restoredRepos := map[string]Repository{
		"first":  storage.New[string, []byte](),
		"second": storage.New[string, []byte](),
	}

	for i := range 1000 {
		idx := strconv.Itoa(i)
		repos["first"].Set("first-"+idx, []byte("first-"+idx))
		repos["second"].Set("second-"+idx, []byte("second-"+idx))
	}

	for i := 0; i < b.N; i++ {
		err1 := dump(gob.NewEncoder(buf), repos)
		err2 := restore(gob.NewDecoder(buf), restoredRepos)
		b.StopTimer()
		if err1 != nil {
			b.Fatal(err1)
		}
		if err2 != nil {
			b.Fatal(err2)
		}
		buf.Reset()
		b.StartTimer()
	}
}
