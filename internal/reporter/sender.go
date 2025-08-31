// Package reporter - sender package.
package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/mailru/easyjson"
	"go.uber.org/multierr"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
	"github.com/ASRafalsky/telemetry/internal/utils"
	pb "github.com/ASRafalsky/telemetry/proto"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

// Config for send data.
type Config struct {
	Key       string         // Key - key for sign data.
	Address   string         // Address - dst address of data.
	RateLimit int            // RateLimit - rate limit for send data.
	Interval  time.Duration  // Interval - period send data.
	PubKey    *rsa.PublicKey // PubKey - public key for encryption.
	ClientIP  string         // ClientIP Local IP address of the client host.
	GRPC      bool
}

// Send sends data from repo with mType through client to the dst from cfg.
func Send(ctx context.Context, mType string, cfg Config, client interface{}, repo repository, log logger) {
	log.Info("Reporeter started with interval:", cfg.Interval.String())
	log.Info("Reporeter started with rate limit:", strconv.Itoa(cfg.RateLimit))
	if len(cfg.ClientIP) != 0 {
		log.Info("Reporeter addr:", cfg.ClientIP)
	} else {
		log.Warn("Reporeter addr undefined")
	}

	sendTimer := time.NewTicker(cfg.Interval)
	defer sendTimer.Stop()

	limit := rate.Inf
	if cfg.RateLimit > 0 {
		limit = rate.Every(time.Second / time.Duration(cfg.RateLimit))
	}

	rl := rate.NewLimiter(limit, 1)
	for {
		select {
		case <-ctx.Done():
			log.Info("Reporeter shutting down")
			gracefulShutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if cfg.GRPC {
				processAndSendGRPC(gracefulShutdownContext, mType, cfg, client.(pb.MetricsClient), repo, log)
			} else {
				processAndSend(gracefulShutdownContext, mType, cfg, client.(*httpclient.Client), repo, log)
			}
			cancel()
			log.Info("Reporeter shutdown complete due to", gracefulShutdownContext.Err().Error())
			return
		case <-sendTimer.C:
			if !rl.Allow() {
				continue
			}
			if cfg.GRPC {
				processAndSendGRPC(ctx, mType, cfg, client.(pb.MetricsClient), repo, log)
			} else {
				processAndSend(ctx, mType, cfg, client.(*httpclient.Client), repo, log)
			}
		}
	}
}

func processAndSend(ctx context.Context, mType string, cfg Config, client *httpclient.Client, repo repository, log logger) {
	sendCtx, cancel := context.WithTimeout(ctx, cfg.Interval*90/100) // I'm so sorry)).

	var bufToSend = bytes.NewBuffer(nil)

	zw := gzip.NewWriter(bufToSend)
	err := serializeMetrics(ctx, "", repo, zw)
	if errZw := zw.Close(); errZw != nil || err != nil {
		err = multierr.Append(err, fmt.Errorf("failed to close gzip writer: %w", errZw))
		log.Error("[send/json] failed to serialize metrics for", mType, err.Error())
		cancel()
		return
	}
	header := http.Header{
		"Content-Type": []string{"application/json"},
		"X-Real-Ip":    []string{cfg.ClientIP}, // I find this strange.
	}
	header.Set("Content-Encoding", "gzip")

	bytesToSend, errEnc := utils.Encrypt(cfg.PubKey, bufToSend.Bytes())
	if errEnc != nil {
		err = multierr.Append(err, fmt.Errorf("failed to encrypt data: %w", errEnc))
		log.Error("[send/json] failed to encrypt data for", mType, err.Error())
		cancel()
		return
	}

	if signature, err := utils.Sign(bytesToSend, []byte(cfg.Key)); err != nil {
		log.Error("[send/json] failed to sign data for", mType, err.Error())
		cancel()
		return
	} else {
		header.Set("HashSHA256", signature)
	}

	if err := withRetryOnErr(sendCtx, 3, func() error {
		return sendDataTo("/updates/", cfg, header, bytes.NewReader(bytesToSend), client)
	}); err != nil {
		log.Error("[send/json] failed to send data] for", mType, err.Error())
	}
	cancel()
}

func processAndSendGRPC(ctx context.Context, mType string, cfg Config, client pb.MetricsClient, repo repository, log logger) {
	sendCtx, cancel := context.WithTimeout(ctx, cfg.Interval*90/100) // I'm so sorry)).

	metricsMap, err := metricsToMap(ctx, mType, repo)
	if err != nil {
		log.Error("[send/grpc] failed to serialize metrics for", mType, err.Error())
		cancel()
		return
	}

	data := pb.MetricsList{
		Values: metricsMap,
	}
	buf, err := proto.Marshal(&data)
	if err != nil {
		log.Error("[send/grpc] failed to serialize metrics for", mType, err.Error())
	}

	md := metadata.New(map[string]string{})
	md.Set("X-Real-Ip", cfg.ClientIP)

	bytesToSend, errEnc := utils.Encrypt(cfg.PubKey, buf)
	if errEnc != nil {
		err = multierr.Append(err, fmt.Errorf("failed to encrypt data: %w", errEnc))
		log.Error("[send/grpc] failed to encrypt data for", mType, err.Error())
		cancel()
		return
	}
	if signature, err := utils.Sign(bytesToSend, []byte(cfg.Key)); err != nil {
		log.Error("[send/grpc] failed to sign data for", mType, err.Error())
		cancel()
		return
	} else {
		md.Set("HashSHA256", signature)
	}
	sendCtx = metadata.NewOutgoingContext(sendCtx, md)

	if err := withRetryOnErr(sendCtx, 3, func() error {
		_, err := client.Set(sendCtx, &pb.SetMetricsRequest{Values: bytesToSend})
		return err
	}); err != nil {
		log.Error("[send/grpc] failed to send data] for", mType, err.Error())
	}
	cancel()
}

func sendDataTo(dst string, cfg Config, header http.Header, r io.Reader, client *httpclient.Client) (err error) {
	resp, err := client.Post(cfg.Address+dst, r, header)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			err = multierr.Append(err, fmt.Errorf("failed to close response body: %w", errClose))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err = multierr.Append(err, fmt.Errorf("bad status: %s", resp.Status))
	}
	return err
}

func sendCounterData(ctx context.Context, addr string, repo repository, client *httpclient.Client) error {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	var errRes error
	return repo.ForEach(ctx, func(k string, v []byte) error {
		if !strings.HasPrefix(k, counter) {
			return nil
		}
		urlData := addr + "/update/counter/" + strings.TrimPrefix(k, counter) + "/" + types.BytesToCounter(v).String()
		resp, err := client.Post(urlData, nil, header)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to send data for %s; %w", k, err))
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			errRes = multierr.Append(errRes, fmt.Errorf("bad response status for %s; %s", k, resp.Status))
		}
		if errClose := resp.Body.Close(); errClose != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to close response body for %s; %w", k, err))
		}
		return errRes
	})
}

func sendGaugeData(ctx context.Context, addr string, repo repository, client *httpclient.Client) error {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}

	var errRes error
	return repo.ForEach(ctx, func(k string, v []byte) error {
		if !strings.HasPrefix(k, gauge) {
			return nil
		}
		urlData := addr + "/update/gauge/" + strings.TrimPrefix(k, gauge) + "/" + types.BytesToGauge(v).String()
		resp, err := client.Post(urlData, nil, header)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to send data for %s; %w", k, err))
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			errRes = multierr.Append(errRes, fmt.Errorf("bad response status for %s; %s", k, resp.Status))
		}
		if errClose := resp.Body.Close(); errClose != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to close response body for %s; %w", k, err))
		}
		return errRes
	})
}

func metricsToMap(ctx context.Context, mtype string, repo repository) (map[string][]byte, error) {
	res := make(map[string][]byte, repo.Size())
	var errRes error
	_ = repo.DropFn(ctx, func(k string, v []byte) (bool, error) {
		typeToSend, key, drop, err := parseKey(k)
		if err != nil {
			return false, nil
		}
		if mtype != "" && mtype != typeToSend {
			return false, nil
		}

		metric, err := dataToMetrics(typeToSend, key, v)
		if err != nil {
			errRes = multierr.Append(errRes,
				fmt.Errorf("failed converting data to metric for %s(%s); %w", mtype, k, err))
			return false, nil
		}
		buf, err := easyjson.Marshal(metric)
		if err != nil {
			errRes = multierr.Append(errRes, err)
			return false, nil
		}
		res[k] = buf
		return drop, nil
	})
	if errRes != nil {
		return nil, errRes
	}
	return res, nil
}

func serializeMetrics(ctx context.Context, mtype string, repo repository, wc writerCloser) error {
	var errRes error
	_ = repo.DropFn(ctx, func(k string, v []byte) (bool, error) {
		typeToSend, key, drop, err := parseKey(k)
		if err != nil {
			return false, nil
		}
		if mtype != "" && mtype != typeToSend {
			return false, nil
		}

		metric, err := dataToMetrics(typeToSend, key, v)
		if err != nil {
			errRes = multierr.Append(errRes,
				fmt.Errorf("failed converting data to metric for %s(%s); %w", mtype, k, err))
			return false, nil
		}
		if err := transport.SerializeMetrics(&metric, wc); err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to compress data for %s(%s); %w", mtype, k, err))
			return false, nil
		}
		return drop, nil
	})
	return errRes
}

func parseKey(k string) (mtype string, id string, drop bool, err error) {
	switch {
	case strings.HasPrefix(k, gauge):
		id = strings.TrimPrefix(k, gauge)
		mtype = gauge
		// Drop this entry, because we will send it and have a new value every time.
		drop = true
	case strings.HasPrefix(k, counter):
		// Do not drop it because we need this value on the next pool call.
		id = strings.TrimPrefix(k, counter)
		mtype = counter
	default:
		return "", "", false, errors.New("unknown metric type")
	}
	return mtype, id, drop, nil
}

func dataToMetrics(mtype, name string, d []byte) (transport.Metrics, error) {
	metrics := transport.Metrics{
		ID:    name,
		MType: mtype,
	}
	switch mtype {
	case counter:
		value := int64(types.BytesToCounter(d))
		metrics.Delta = &value
	case gauge:
		value := float64(types.BytesToGauge(d))
		metrics.Value = &value
	default:
		return transport.Metrics{}, fmt.Errorf("unknown metrics type: %s", mtype)
	}
	return metrics, nil
}

func withRetryOnErr(ctx context.Context, cnt int, fn func() error) error {
	cnt--
	err := fn()
	if err != nil && (errors.Is(err, syscall.ECONNREFUSED) ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "unreachable") ||
		strings.Contains(err.Error(), "no route to host") ||
		strings.Contains(err.Error(), "invalid argument")) {
		wait := 1
		ticker := time.NewTicker(time.Duration(wait) * time.Second)
		defer ticker.Stop()
		for cnt >= 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				err = fn()
				if err == nil {
					return nil
				}
				wait += 2
				ticker.Reset(time.Duration(wait) * time.Second)
				cnt--
			}
		}
	}
	return err
}

type logger interface {
	Info(msg string, add ...string)
	Warn(msg string, add ...string)
	Error(msg string, add ...string)
	Debug(msg string, add ...string)
	Fatal(msg string, add ...string)
}

type repository interface {
	DropFn(ctx context.Context, fn func(k string, v []byte) (bool, error)) error
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
}

type writerCloser interface {
	Write(p []byte) (n int, err error)
	Close() error
}
