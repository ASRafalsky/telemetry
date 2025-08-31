package main

import (
	"fmt"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/utils"
	"github.com/ASRafalsky/telemetry/pkg/log"
	pb "github.com/ASRafalsky/telemetry/proto"
)

// Non-production code. Please, use something more secure.
// TODO(ASRafalsky): Do something about insecure option.
func newClient(cfg config.Agent, logger *log.Logger) (interface{}, string) {
	laddr, err := utils.GetAddr("tcp")
	if err != nil {
		logger.Error("Failed to get client IP:", err.Error())
	}
	if cfg.GRPC {
		conn, err := grpc.NewClient(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Errorf("failed to create new grpc client: %w", err))
		}
		return pb.NewMetricsClient(conn), laddr
	}
	// Create a new HTTP client with a default timeout
	timeout := 10 * time.Second
	return httpclient.NewClient(httpclient.WithHTTPTimeout(timeout)), laddr
}
