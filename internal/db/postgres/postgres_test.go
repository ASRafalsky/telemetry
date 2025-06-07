package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/transport"
)

func TestDB(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationFilesPath := filepath.Dir(file) + "/test_migrations"

	db, err := Open("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Skip("Failed to connect to database: ", err.Error())
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Log("Failed to close db:", err.Error())
		}
	}()

	ctx := context.Background()
	if err = db.WaitDBIsReady(ctx, 5, time.Second); err != nil {
		t.Skip("Failed to connect to database: ", err.Error())
	}

	const (
		cnt           = 1000
		insertData    = `INSERT INTO metrics_test (id, payload) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET payload = $2`
		selectData    = `SELECT payload FROM metrics_test WHERE id = $1`
		deleteData    = `DELETE FROM metrics_test WHERE id = $1`
		selectForEach = `SELECT id, payload FROM metrics_test ORDER BY id LIMIT $1 OFFSET $2`
		insertBatch   = `INSERT INTO metrics_test (id, payload) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET payload = $2`
	)

	if err = db.MigrateUp(migrationFilesPath); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			require.NoError(t, err)
		}
	}

	metricsList := make([]transport.Metrics, cnt)
	keys := make([]string, cnt)
	for i := range cnt {
		var (
			val   = rand.Float64()
			delta = rand.Int64()
		)
		metricsList[i] = transport.Metrics{
			MType: "MType" + strconv.Itoa(i),
			ID:    "ID" + strconv.Itoa(i),
			Value: &val,
			Delta: &delta,
		}
		keys[i] = metricsList[i].MType + metricsList[i].ID
	}

	t.Run("Set", func(t *testing.T) {
		for i := range metricsList {
			buf, err := json.Marshal(metricsList[i])
			require.NoError(t, err)
			require.NoError(t, db.set(ctx, insertData, keys[i], buf))
		}
	})

	t.Run("Get", func(t *testing.T) {
		for i := range metricsList {
			res, err := db.get(ctx, selectData, keys[i])
			require.NoError(t, err)
			var m transport.Metrics
			require.NoError(t, json.Unmarshal(res, &m))
			require.Equal(t, metricsList[i], m)
		}
	})

	t.Run("Update", func(t *testing.T) {
		for i := range metricsList {
			buf, err := json.Marshal(metricsList[i])
			require.NoError(t, err)
			require.NoError(t, db.set(ctx, insertData, keys[i], buf))
		}
	})

	t.Run("ForEach", func(t *testing.T) {
		require.NoError(t, db.forEach(ctx, selectForEach, func(k string, v []byte) error {
			require.Contains(t, keys, k)
			return nil
		}))
	})

	t.Run("SetBatch", func(t *testing.T) {
		_, err = db.ExecContext(ctx, `DELETE FROM metrics_test`)
		require.NoError(t, err)
		pairs := make([]Pair, cnt)
		for i := range metricsList {
			buf, err := json.Marshal(metricsList[i])
			require.NoError(t, err)
			pairs[i] = Pair{
				ID:      keys[i],
				Payload: buf,
			}
		}
		require.NoError(t, db.setBatch(ctx, insertBatch, pairs))

		for i := range metricsList {
			res, err := db.get(ctx, selectData, keys[i])
			require.NoError(t, err)
			var m transport.Metrics
			require.NoError(t, json.Unmarshal(res, &m))
			require.Equal(t, metricsList[i], m)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		for i := range keys {
			res, err := db.get(ctx, selectData, keys[i])
			require.NoError(t, err)
			require.NotNil(t, res)
			err = db.deleteEntrie(ctx, deleteData, keys[i])
			require.NoError(t, err)
			res, err = db.get(ctx, selectData, keys[i])
			require.NoError(t, err)
			require.Nil(t, res)
		}
	})
}
