package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mailru/easyjson"

	"github.com/ASRafalsky/telemetry/internal/transport"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func counterPostDataHandler(repo repository, value transport.Metrics) ([]byte, error) {
	name := Counter + value.ID
	if buf, ok := repo.Get(name); ok {
		previousValue := transport.Metrics{}
		if err := easyjson.Unmarshal(buf, &previousValue); err != nil {
			return nil, err
		}
		*value.Delta += *previousValue.Delta
	}
	buf, err := easyjson.Marshal(&value)
	if err != nil {
		return nil, err
	}
	repo.Set(name, buf)
	return buf, nil
}

func gaugeGetDataHandler(repo repository, key string) ([]byte, error) {
	if buf, ok := repo.Get(Gauge + key); ok {
		return buf, nil
	}
	return nil, errors.New("gauge value not found")
}

func counterGetDataHandler(repo repository, key string) ([]byte, error) {
	if buf, ok := repo.Get(Counter + key); ok {
		return buf, nil
	}
	return nil, errors.New("counter value not found")
}

func gaugePostDataHandler(repo repository, value transport.Metrics) ([]byte, error) {
	buf, err := easyjson.Marshal(&value)
	if err != nil {
		return nil, err
	}
	repo.Set(Gauge+value.ID, buf)
	return buf, nil
}

func SetDataTo(repo repository, m transport.Metrics) ([]byte, int, error) {
	switch m.MType {
	case Gauge:
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
	case Counter:
		switch {
		case m.Delta != nil:
			dataBuf, err := counterPostDataHandler(repo, m)
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

func GetDataFrom(repo repository, m transport.Metrics) ([]byte, int, error) {
	switch m.MType {
	case Gauge:
		dataBuf, err := gaugeGetDataHandler(repo, m.ID)
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		return dataBuf, http.StatusOK, nil
	case Counter:
		dataBuf, err := counterGetDataHandler(repo, m.ID)
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		return dataBuf, http.StatusOK, nil
	default:
		return nil, http.StatusBadRequest, fmt.Errorf("type %s not supported", m.MType)
	}
}

func getKeyList(repo repository) []string {
	result := make([]string, repo.Size())
	_ = repo.ForEach(context.Background(), func(k string, _ []byte) error {
		switch {
		case strings.HasPrefix(k, Gauge):
			result = append(result, strings.TrimPrefix(k, Gauge))
		case strings.HasPrefix(k, Counter):
			result = append(result, strings.TrimPrefix(k, Counter))
		default:
			return nil
		}
		return nil
	})
	return result
}
