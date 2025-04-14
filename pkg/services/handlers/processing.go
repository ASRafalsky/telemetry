package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func counterPostDataHandler(repo Repository, key string, value types.Counter) (types.Counter, error) {
	if previousValue, ok := repo.Get(strings.ToLower(key)); ok {
		value += types.BytesToCounter(previousValue)
	}
	repo.Set(strings.ToLower(key), types.CounterToBytes(value))
	return value, nil
}

func gaugeGetDataHandler(repo Repository, key string) (types.Gauge, error) {
	if value, ok := repo.Get(strings.ToLower(key)); ok {
		return types.BytesToGauge(value), nil
	}
	return types.Gauge(0), errors.New("gauge value not found")
}

func counterGetDataHandler(repo Repository, key string) (types.Counter, error) {
	if value, ok := repo.Get(strings.ToLower(key)); ok {
		return types.BytesToCounter(value), nil
	}
	return types.Counter(0), errors.New("counter value not found")
}

func gaugePostDataHandler(repo Repository, key string, value types.Gauge) (types.Gauge, error) {
	repo.Set(strings.ToLower(key), types.GaugeToBytes(value))
	return value, nil
}

func SetDataTo(repo Repository, m transport.Metrics) (transport.Metrics, int, error) {
	switch m.MType {
	case Gauge:
		switch {
		case m.Value != nil:
			value, err := gaugePostDataHandler(repo, m.ID, types.Gauge(*m.Value))
			if err != nil {
				return transport.Metrics{}, http.StatusInternalServerError, err
			}
			valFloat := float64(value)
			m.Value = &valFloat
			return m, http.StatusOK, nil
		case m.Delta != nil:
			return transport.Metrics{}, http.StatusBadRequest, errors.New("delta not supported for gauge")
		default:
			return transport.Metrics{}, http.StatusBadRequest, errors.New("gaugePostDataHandler called with no data")
		}
	case Counter:
		switch {
		case m.Delta != nil:
			value, err := counterPostDataHandler(repo, m.ID, types.Counter(*m.Delta))
			if err != nil {
				return transport.Metrics{}, http.StatusInternalServerError, err
			}
			valInt64 := int64(value)
			m.Delta = &valInt64
			return m, http.StatusOK, nil
		case m.Value != nil:
			return transport.Metrics{}, http.StatusBadRequest, errors.New("value not supported for gauge")
		default:
			return transport.Metrics{}, http.StatusBadRequest, errors.New("gaugePostDataHandler called with no data")
		}
	default:
		return transport.Metrics{}, http.StatusBadRequest, fmt.Errorf("type %s not supported", m.MType)
	}
}

func GetDataFrom(repo Repository, m transport.Metrics) (transport.Metrics, int, error) {
	switch m.MType {
	case Gauge:
		value, err := gaugeGetDataHandler(repo, m.ID)
		if err != nil {
			return transport.Metrics{}, http.StatusNotFound, err
		}
		valFloat := float64(value)
		m.Value = &valFloat
		return m, http.StatusOK, nil
	case Counter:
		value, err := counterGetDataHandler(repo, m.ID)
		if err != nil {
			return transport.Metrics{}, http.StatusNotFound, err
		}
		valInt64 := int64(value)
		m.Delta = &valInt64
		return m, http.StatusOK, nil
	default:
		return transport.Metrics{}, http.StatusBadRequest, fmt.Errorf("type %s not supported", m.MType)
	}
}

func getKeyList(repos map[string]Repository) []string {
	totalEntryCnt := 0
	for _, repo := range repos {
		totalEntryCnt += repo.Size()
	}

	result := make([]string, totalEntryCnt)
	for _, repo := range repos {
		_ = repo.ForEach(context.Background(), func(k string, _ []byte) error {
			result = append(result, k)
			return nil
		})
	}
	return result
}
