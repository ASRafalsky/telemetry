package interceptor

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/ASRafalsky/telemetry/internal/handlers"
	testrepo "github.com/ASRafalsky/telemetry/internal/repository"
	"github.com/ASRafalsky/telemetry/internal/utils"
	"github.com/ASRafalsky/telemetry/pkg/cache"
	"github.com/ASRafalsky/telemetry/pkg/log"
	pb "github.com/ASRafalsky/telemetry/proto"
)

const (
	secretKey            = "really secret key"
	someAnotherSecretKey = "ololo"
)

func TestSignInterceptor(t *testing.T) {
	repo := testrepo.NewExtendedRepository(cache.New[string, []byte]())
	srv := grpc.NewServer(grpc.UnaryInterceptor(NewSignInterceptor([]byte(secretKey))))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	listen, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	go func() {
		defer wg.Done()
		pb.RegisterMetricsServer(srv, handlers.MetricsServer{Repo: repo})
		require.NoError(t, srv.Serve(listen))
	}()

	conn, err := grpc.NewClient(listen.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() {
		require.NoError(t, conn.Close())
	}()

	mc := pb.NewMetricsClient(conn)

	mMap, _, _ := utils.MetricsMap(t)
	data := pb.MetricsList{
		Values: mMap,
	}
	buf, err := proto.Marshal(&data)
	require.NoError(t, err)

	t.Run("Signature OK", func(t *testing.T) {
		signature, err := utils.Sign(buf, []byte(secretKey))
		require.NoError(t, err)
		md := metadata.New(map[string]string{})
		md.Set("HashSHA256", signature)
		respSet, err := mc.Set(metadata.NewOutgoingContext(ctx, md), &pb.SetMetricsRequest{Values: buf})
		require.NoError(t, err)
		require.Empty(t, respSet.Error)
	})

	t.Run("Signature NOK", func(t *testing.T) {
		signature, err := utils.Sign(buf, []byte(someAnotherSecretKey))
		require.NoError(t, err)
		md := metadata.New(map[string]string{})
		md.Set("HashSHA256", signature)
		_, err = mc.Set(metadata.NewOutgoingContext(ctx, md), &pb.SetMetricsRequest{Values: buf})
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid signature")
	})

	t.Run("Signature Empty", func(t *testing.T) {
		_, err = mc.Set(ctx, &pb.SetMetricsRequest{Values: buf})
		require.Error(t, err)
		require.ErrorContains(t, err, "no required signature")
	})

	cancel()
	srv.Stop()
	wg.Wait()
}

func TestDecryptInterceptor(t *testing.T) {
	repo := testrepo.NewExtendedRepository(cache.New[string, []byte]())
	privateKey, err := utils.ParsePrivateKey("./testdata/private.key")
	require.NoError(t, err)
	srv := grpc.NewServer(grpc.UnaryInterceptor(NewDecryptInterceptor(privateKey)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	listen, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	go func() {
		defer wg.Done()
		pb.RegisterMetricsServer(srv, handlers.MetricsServer{Repo: repo})
		require.NoError(t, srv.Serve(listen))
	}()

	conn, err := grpc.NewClient(listen.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() {
		require.NoError(t, conn.Close())
	}()
	mc := pb.NewMetricsClient(conn)

	mMap, _, _ := utils.MetricsMap(t)
	data := pb.MetricsList{
		Values: mMap,
	}
	buf, err := proto.Marshal(&data)
	require.NoError(t, err)

	t.Run("Decrypt OK", func(t *testing.T) {
		publicKey, err := utils.ParsePublicKey("./testdata/public.key")
		require.NoError(t, err)
		encriptedData, err := utils.Encrypt(publicKey, buf)
		require.NoError(t, err)
		sz, err := repo.Size()
		require.NoError(t, err)
		require.Equal(t, sz, 0)
		respSet, err := mc.Set(ctx, &pb.SetMetricsRequest{Values: encriptedData})
		require.NoError(t, err)
		require.Empty(t, respSet.Error)
		sz, err = repo.Size()
		require.NoError(t, err)
		require.Equal(t, sz, len(mMap))
	})

	t.Run("Decrypt NOK", func(t *testing.T) {
		_, err := mc.Set(ctx, &pb.SetMetricsRequest{Values: buf})
		require.ErrorContains(t, err, "decryption error")
	})
}

func TestInterceptorsChain(t *testing.T) {
	Log, err := log.AddLoggerWith("Info", "")
	require.NoError(t, err)
	defer Log.Sync()
	repo := testrepo.NewExtendedRepository(cache.New[string, []byte]())
	privateKey, err := utils.ParsePrivateKey("./testdata/private.key")
	require.NoError(t, err)
	addr, err := utils.GetAddr("tcp")
	require.NoError(t, err)
	_, cidr, err := net.ParseCIDR(addr + "/24")
	require.NoError(t, err)
	srv := grpc.NewServer(grpc.UnaryInterceptor(NewInterceptorsChain(true, cidr, []byte(secretKey), privateKey, Log)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	listen, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	go func() {
		defer wg.Done()
		pb.RegisterMetricsServer(srv, handlers.MetricsServer{Repo: repo})
		require.NoError(t, srv.Serve(listen))
	}()

	conn, err := grpc.NewClient(listen.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() {
		require.NoError(t, conn.Close())
	}()
	mc := pb.NewMetricsClient(conn)

	mMap, _, _ := utils.MetricsMap(t)
	data := pb.MetricsList{
		Values: mMap,
	}
	buf, err := proto.Marshal(&data)
	require.NoError(t, err)

	t.Run("OOOOK!!!", func(t *testing.T) {
		sz, err := repo.Size()
		require.NoError(t, err)
		require.Equal(t, sz, 0)
		publicKey, err := utils.ParsePublicKey("./testdata/public.key")
		require.NoError(t, err)
		encryptedData, err := utils.Encrypt(publicKey, buf)
		require.NoError(t, err)
		signature, err := utils.Sign(encryptedData, []byte(secretKey))
		require.NoError(t, err)
		md := metadata.New(map[string]string{})
		md.Set("HashSHA256", signature)
		md.Set("X-Real-Ip", addr)
		respSet, err := mc.Set(metadata.NewOutgoingContext(ctx, md), &pb.SetMetricsRequest{Values: encryptedData})
		require.NoError(t, err)
		require.Empty(t, respSet.Error)
		sz, err = repo.Size()
		require.NoError(t, err)
		require.Equal(t, sz, len(mMap))
	})
}
