package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/internal/transport"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

func TestBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "/test/backup")

	repo := storage.New[string, []byte]()

	for i := range 100 {
		idx := strconv.Itoa(i)
		gaugeVal := float64(i)
		counterVal := int64(i)
		buf := bytes.NewBuffer([]byte{})
		require.NoError(t, transport.SerializeMetrics(&transport.Metrics{
			MType: gauge,
			ID:    idx,
			Value: &gaugeVal,
		}, buf))
		repo.Set(gauge+idx, buf.Bytes())
		buf = bytes.NewBuffer([]byte{})
		require.NoError(t, transport.SerializeMetrics(&transport.Metrics{
			MType: counter,
			ID:    idx,
			Delta: &counterVal,
		}, buf))
		repo.Set(counter+idx, buf.Bytes())
	}

	// Add data to dump.
	require.NoError(t, DumpRepoToFile(path, repo, 0o644))

	// Check the dump file.
	stat, err := os.Stat(path)
	require.NoError(t, err)
	require.NotZero(t, stat.Size())

	restoredRepo := storage.New[string, []byte]()

	// Try to restore from the dump file.
	err = RestoreRepoFromFile(path, restoredRepo, false)
	require.NoError(t, err)
	require.Equal(t, repo.Size(), restoredRepo.Size())
	for i := range 100 {
		idx := strconv.Itoa(i)
		gaugeVal := float64(i)
		counterVal := int64(i)
		buf, ok := restoredRepo.Get(gauge + idx)
		m, err := transport.DeserializeMetrics(buf)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, gauge, m[0].MType)
		require.Equal(t, idx, m[0].ID)
		require.NotNil(t, m[0].Value)
		require.Equal(t, gaugeVal, *m[0].Value)
		require.Nil(t, m[0].Delta)

		buf, ok = restoredRepo.Get(counter + idx)
		require.True(t, ok)
		m, err = transport.DeserializeMetrics(buf)
		require.NoError(t, err)
		require.Equal(t, counter, m[0].MType)
		require.Equal(t, idx, m[0].ID)
		require.NotNil(t, m[0].Delta)
		require.Equal(t, counterVal, *m[0].Delta)
		require.Nil(t, m[0].Value)
	}

	restoredRepo2 := storage.New[string, []byte]()

	err = RestoreRepoFromFile(path, restoredRepo2, true)
	require.NoError(t, err)
	for i := range 100 {
		idx := strconv.Itoa(i)
		gaugeVal := float64(i)
		counterVal := int64(i)
		buf, ok := restoredRepo.Get(gauge + idx)
		m, err := transport.DeserializeMetrics(buf)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, gauge, m[0].MType)
		require.Equal(t, idx, m[0].ID)
		require.NotNil(t, m[0].Value)
		require.Equal(t, gaugeVal, *m[0].Value)
		require.Nil(t, m[0].Delta)

		buf, ok = restoredRepo.Get(counter + idx)
		require.True(t, ok)
		m, err = transport.DeserializeMetrics(buf)
		require.NoError(t, err)
		require.Equal(t, counter, m[0].MType)
		require.Equal(t, idx, m[0].ID)
		require.NotNil(t, m[0].Delta)
		require.Equal(t, counterVal, *m[0].Delta)
		require.Nil(t, m[0].Value)
	}
	// Yes, you are doing everything right.)))
	_, err = os.Stat(path)
	require.Error(t, err, os.ErrNotExist)
}
