package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mailru/easyjson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
	pb "github.com/ASRafalsky/telemetry/proto"
)

// MetricsServer all required server methods.
type MetricsServer struct {
	pb.UnimplementedMetricsServer
	Repo repository
}

// MetricsMultiServer all required server methods.
type MetricsMultiServer struct {
	pb.UnimplementedMetricsMultiServer
	Repo repository
}

type DBServer struct {
	pb.UnimplementedBDServer
	Repo repository
}

// Set sets data.
func (s MetricsServer) Set(ctx context.Context, in *pb.SetMetricsRequest) (*pb.SetMetricsResponse, error) {
	var response pb.SetMetricsResponse
	if len(in.Values) == 0 {
		response.Error = "empty values"
		return &response, status.Error(codes.InvalidArgument, "empty values")
	}

	var data pb.MetricsList
	if err := proto.Unmarshal(in.Values, &data); err != nil {
		response.Error = err.Error()
		return &response, status.Errorf(codes.Internal, "ummarshal error: %v", err)
	}
	gaugeData := make(map[string][]byte, len(in.Values))
	for k, v := range data.Values {
		if strings.HasPrefix(k, types.GaugeType) {
			gaugeData[k] = v
			delete(data.Values, k)
		}
	}
	if len(gaugeData) > 0 {
		s.Repo.Merge(gaugeData)
	}

	for k, v := range data.Values {
		if !strings.HasPrefix(k, types.CounterType) {
			response.Error = fmt.Sprintf("unknown metric type: %s", k)
			return &response, status.Errorf(codes.InvalidArgument, "unknown metric type: %s", k)
		}
		var m transport.Metrics
		if err := easyjson.Unmarshal(v, &m); err != nil {
			response.Error = err.Error()
			return &response, status.Errorf(codes.Internal, "metric ummarshal error: %v", err)
		}
		if m.Delta == nil {
			continue
		}
		_, err := counterDataHandler(ctx, s.Repo, m)
		if err != nil {
			response.Error = err.Error()
			return &response, status.Errorf(codes.Internal, "set data error: %v", err)
		}
	}
	return &response, nil
}

// Get returns data.
func (s MetricsMultiServer) Get(in *pb.GetMetricsRequest, srv pb.MetricsMulti_GetServer) error {
	if len(in.Id) == 0 {
		return errors.New("empty list requested")
	}

	for _, id := range in.Id {
		var (
			m     transport.Metrics
			mtype string
			value []byte
		)
		switch in.Mtype {
		case pb.Mtype_GAUGE:
			mtype = types.GaugeType
		case pb.Mtype_COUNTER:
			mtype = types.CounterType
		default:
			return status.Errorf(codes.InvalidArgument, "unsupported type: %d", in.Mtype)
		}

		buf, err := s.Repo.Get(srv.Context(), mtype+id)
		if err != nil {
			return status.Errorf(codes.Internal, "get data error: %v", err)
		}
		if len(buf) == 0 {
			continue
		}

		if err = easyjson.Unmarshal(buf, &m); err != nil {
			return status.Errorf(codes.Internal, "unmarshal error: %v", err)
		}

		if m.Value != nil {
			value = types.GaugeToBytes(types.Gauge(*m.Value))
		}
		if len(value) == 0 && m.Delta != nil {
			value = types.CounterToBytes(types.Counter(*m.Delta))
		}
		if len(value) == 0 {
			return status.Errorf(codes.NotFound, "metric not found: %s", id)
		}

		err = srv.Send(&pb.GetMetricsResponse{
			Mtype: in.Mtype,
			Id:    id,
			Value: value,
		})
		if err != nil {
			return status.Errorf(codes.Internal, "send data error: %v", err)
		}
	}
	return nil
}

func (s DBServer) Ping(ctx context.Context, _ *pb.BDPingRequest) (*pb.BDPingResponse, error) {
	ctxPing, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := s.Repo.Ping(ctxPing); err != nil {
		return &pb.BDPingResponse{Error: err.Error()}, nil
	}
	return &pb.BDPingResponse{}, nil
}
