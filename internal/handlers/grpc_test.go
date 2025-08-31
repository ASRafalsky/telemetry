package handlers

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/protobuf/proto"

	testrepo "github.com/ASRafalsky/telemetry/internal/repository"
	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
	"github.com/ASRafalsky/telemetry/internal/utils"
	"github.com/ASRafalsky/telemetry/pkg/cache"
	pb "github.com/ASRafalsky/telemetry/proto"
)

func TestGRPCServer(t *testing.T) {
	repo := testrepo.NewExtendedRepository(cache.New[string, []byte]())
	srv := grpc.NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		listen, err := net.Listen("tcp", ":3200")
		require.NoError(t, err)
		pb.RegisterMetricsServer(srv, MetricsServer{Repo: repo})
		pb.RegisterMetricsMultiServer(srv, MetricsMultiServer{Repo: repo})
		require.NoError(t, srv.Serve(listen))
	}()

	conn, err := grpc.NewClient(
		":3200",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, conn.Close())
	}()

	mc := pb.NewMetricsClient(conn)
	multiMc := pb.NewMetricsMultiClient(conn)

	metricsMap, vKeys, dKeys := utils.MetricsMap(t)

	data := pb.MetricsList{
		Values: metricsMap,
	}
	buf, err := proto.Marshal(&data)
	require.NoError(t, err)

	t.Run("Set", func(t *testing.T) {
		sz, err := repo.Size()
		require.NoError(t, err)
		require.Equal(t, sz, 0)
		respSet, err := mc.Set(ctx, &pb.SetMetricsRequest{Values: buf})
		require.NoError(t, err)
		require.Empty(t, respSet.Error)
		sz, err = repo.Size()
		require.NoError(t, err)
		require.Equal(t, sz, len(metricsMap))
	})

	t.Run("Get_multi_gauge", func(t *testing.T) {
		respGetValues, err := multiMc.Get(ctx, &pb.GetMetricsRequest{
			Mtype: pb.Mtype_GAUGE,
			Id:    vKeys,
		})
		require.NoError(t, err)
		cnt := 0
		for {
			require.NotNil(t, respGetValues)
			resp, err := respGetValues.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			cnt++
			m := transport.Metrics{}
			require.NoError(t, easyjson.Unmarshal(metricsMap[types.GaugeType+resp.Id], &m))
			require.Equal(t, *m.Value, float64(types.BytesToGauge(resp.Value)))
		}
		require.Equal(t, len(vKeys), cnt)
	})

	t.Run("Get_multi_cnt", func(t *testing.T) {
		respGetDeltas, err := multiMc.Get(ctx, &pb.GetMetricsRequest{
			Mtype: pb.Mtype_COUNTER,
			Id:    dKeys,
		})
		require.NoError(t, err)
		cnt := 0
		for {
			require.NotNil(t, respGetDeltas)
			resp, err := respGetDeltas.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			cnt++
			m := transport.Metrics{}
			require.NoError(t, easyjson.Unmarshal(metricsMap[types.CounterType+resp.Id], &m))
			require.Equal(t, *m.Delta, int64(types.BytesToCounter(resp.Value)))
		}
		require.Equal(t, len(dKeys), cnt)
	})

	t.Run("Set + Get", func(t *testing.T) {
		sz, err := repo.Size()
		require.NoError(t, err)
		// It was set before.
		require.Equal(t, sz, len(metricsMap))
		respSet, err := mc.Set(ctx, &pb.SetMetricsRequest{Values: buf})
		require.NoError(t, err)
		require.Empty(t, respSet.Error)
		sz, err = repo.Size()
		require.NoError(t, err)
		require.Equal(t, sz, len(metricsMap))

		respGetDeltas, err := multiMc.Get(ctx, &pb.GetMetricsRequest{
			Mtype: pb.Mtype_COUNTER,
			Id:    dKeys,
		})
		require.NoError(t, err)
		cnt := 0
		for {
			require.NotNil(t, respGetDeltas)
			resp, err := respGetDeltas.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			cnt++
			m := transport.Metrics{}
			require.NoError(t, easyjson.Unmarshal(metricsMap[types.CounterType+resp.Id], &m))
			// resp.Value is sum of previous value with the same value.
			require.Equal(t, *m.Delta*2, int64(types.BytesToCounter(resp.Value)))
		}
		require.Equal(t, len(dKeys), cnt)
	})

	t.Run("Set_invalid_metrics", func(t *testing.T) {
		data := pb.MetricsList{
			Values: map[string][]byte{
				"lol": []byte("lol"),
			},
		}
		buf, err := proto.Marshal(&data)
		require.NoError(t, err)
		respSet, err := mc.Set(ctx, &pb.SetMetricsRequest{Values: buf})
		require.Contains(t, err.Error(), "unknown metric type:")
		require.Nil(t, respSet)
	})

	cancel()
	srv.Stop()
	wg.Wait()
}
