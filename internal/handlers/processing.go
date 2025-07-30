package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mailru/easyjson"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

var (
	errCounterNotFound = errors.New("counter value not found")
	errGaugeNotFound   = errors.New("gauge value not found")
)

// SetDataTo sets data to repository and returns serialized metrics, http status code and error if something went wrong.
func SetDataTo(ctx context.Context, repo repository, m transport.Metrics) ([]byte, int, error) {
	switch m.MType {
	case types.GaugeType:
		switch {
		case m.Value != nil:
			dataBuf, err := gaugePostDataHandler(repo, m)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			return dataBuf, http.StatusOK, nil
		case m.Delta != nil:
			return nil, http.StatusBadRequest, errors.New("delta not supported for gauge")
		default:
			return nil, http.StatusBadRequest, errors.New("gaugePostDataHandler called with no data")
		}
	case types.CounterType:
		switch {
		case m.Delta != nil:
			dataBuf, err := counterPostDataHandler(ctx, repo, m)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			return dataBuf, http.StatusOK, nil
		case m.Value != nil:
			return nil, http.StatusBadRequest, errors.New("value not supported for gauge")
		default:
			return nil, http.StatusBadRequest, errors.New("gaugePostDataHandler called with no data")
		}
	default:
		return nil, http.StatusBadRequest, fmt.Errorf("type %s not supported", m.MType)
	}
}

// GetDataFrom receives data from repository and returns serialized metrics, http status code and error,
// if something went wrong.
func GetDataFrom(ctx context.Context, repo repository, m transport.Metrics) ([]byte, int, error) {
	switch m.MType {
	case types.GaugeType:
		dataBuf, err := gaugeGetDataHandler(ctx, repo, m.ID)
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		return dataBuf, http.StatusOK, nil
	case types.CounterType:
		dataBuf, err := counterGetDataHandler(ctx, repo, m.ID)
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		return dataBuf, http.StatusOK, nil
	default:
		return nil, http.StatusBadRequest, fmt.Errorf("type %s not supported", m.MType)
	}
}

func counterPostDataHandler(ctx context.Context, repo repository, value transport.Metrics) ([]byte, error) {
	name := types.CounterType + value.ID
	buf, err := repo.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	if buf != nil {
		previousValue := transport.Metrics{}
		if err = easyjson.Unmarshal(buf, &previousValue); err != nil {
			return nil, err
		}
		*value.Delta += *previousValue.Delta
	}

	buf, err = easyjson.Marshal(&value)
	if err != nil {
		return nil, err
	}
	repo.Set(name, buf)
	return buf, nil
}

func gaugeGetDataHandler(ctx context.Context, repo repository, key string) ([]byte, error) {
	buf, err := repo.Get(ctx, types.GaugeType+key)
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, errGaugeNotFound
	}
	return buf, nil
}

func counterGetDataHandler(ctx context.Context, repo repository, key string) ([]byte, error) {
	buf, err := repo.Get(ctx, types.CounterType+key)
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, errCounterNotFound
	}
	return buf, nil
}

func gaugePostDataHandler(repo repository, value transport.Metrics) ([]byte, error) {
	buf, err := easyjson.Marshal(&value)
	if err != nil {
		return nil, err
	}
	repo.Set(types.GaugeType+value.ID, buf)
	return buf, nil
}

func getKeyList(repo repository) ([]string, error) {
	sz, err := repo.Size()
	if err != nil {
		return nil, err
	}
	result := make([]string, sz)
	_ = repo.ForEach(context.Background(), func(k string, _ []byte) error {
		switch {
		case strings.HasPrefix(k, types.GaugeType):
			result = append(result, strings.TrimPrefix(k, types.GaugeType))
		case strings.HasPrefix(k, types.CounterType):
			result = append(result, strings.TrimPrefix(k, types.CounterType))
		default:
			return nil
		}
		return nil
	})
	return result, nil
}
