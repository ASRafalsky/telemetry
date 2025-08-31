// Package interceptor - all interceptor that we need.
package interceptor

import (
	"context"
	"crypto/rsa"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/ASRafalsky/telemetry/internal/utils"
	"github.com/ASRafalsky/telemetry/pkg/log"
	pb "github.com/ASRafalsky/telemetry/proto"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
)

// NewSignInterceptor checks data signature.
func NewSignInterceptor(signKey []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if len(signKey) == 0 {
			return handler(ctx, req)
		}
		var sign string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("HashSHA256")
			if len(values) == 0 {
				return nil, status.Error(codes.Unauthenticated, "no required signature")
			}
			sign = values[0]
		}
		if r, ok := req.(*pb.SetMetricsRequest); ok {
			if len(r.Values) == 0 {
				return nil, status.Error(codes.Unauthenticated, "invalid signature, empty value")
			}
			ok, err := utils.SignCheck(r.Values, signKey, sign)
			if err == nil && !ok {
				return nil, status.Error(codes.Unauthenticated, "invalid signature")
			}
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, err.Error())
			}
			return handler(ctx, req)
		}
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}
}

// NewSubnetInterceptor checks subnet.
func NewSubnetInterceptor(meta bool, cidr *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cidr == nil {
			return handler(ctx, req)
		}
		var srcIP string
		if meta {
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				values := md.Get("X-Real-IP")
				if len(values) == 0 {
					return nil, status.Error(codes.PermissionDenied, "missing required source IP")
				}
				srcIP = values[0]
			}
		} else {
			p, ok := peer.FromContext(ctx)
			if !ok {
				return nil, status.Error(codes.PermissionDenied, "missing required source IP")
			}
			srcIP = p.Addr.String()
		}
		if srcIP == "" {
			return nil, status.Error(codes.PermissionDenied, "missing source IP")
		}
		if cidr.Contains(net.ParseIP(srcIP)) {
			return handler(ctx, req)
		}
		return nil, status.Error(codes.PermissionDenied, "source IP address is invalid")
	}
}

// NewDecryptInterceptor hmmm...
func NewDecryptInterceptor(key *rsa.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if key == nil {
			return handler(ctx, req)
		}
		if r, ok := req.(*pb.SetMetricsRequest); ok {
			out, err := key.Decrypt(nil, r.Values, nil)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			// Looks bad, but it seems to me ok in this case.
			r.Values = out
			return handler(ctx, r)
		}
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}
}

// NewLoggingInterceptor you can try to guess.
func NewLoggingInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if logger == nil {
			return handler(ctx, req)
		}
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		p, ok := peer.FromContext(ctx)
		var srcIP string
		if ok {
			srcIP = p.Addr.String()
		}

		logger.Info("[gRPC/Request]",
			"url:", srcIP,
			"method:", info.FullMethod,
			"duration:", duration.String(),
		)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				logger.Error("[gRPC/Response/Error]", st.Code().String(), st.Message())
			}
		}
		return resp, err
	}
}

// NewInterceptorsChain It does some of the things that are written in its name.
func NewInterceptorsChain(
	meta bool,
	cidr *net.IPNet,
	signKey []byte,
	key *rsa.PrivateKey,
	logger *log.Logger,
) grpc.UnaryServerInterceptor {
	res := []grpc.UnaryServerInterceptor{NewLoggingInterceptor(logger)}
	// Order is matter.
	if cidr != nil {
		res = append(res, NewSubnetInterceptor(meta, cidr))
	}
	if len(signKey) > 0 {
		res = append(res, NewSignInterceptor(signKey))
	}
	if key != nil {
		res = append(res, NewDecryptInterceptor(key))
	}
	return grpc_middleware.ChainUnaryServer(res...)
}
