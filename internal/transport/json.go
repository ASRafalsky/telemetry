// Package transport describes everything that we need to send and receive metrics.
package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/mailru/easyjson"
)

//go:generate easyjson --all json.go

// Metrics describes transport struct for metric values.
type Metrics struct {
	ID    string   `json:"id"`              // Name of metric.
	MType string   `json:"type"`            // Type of metric, gauge or counter.
	Delta *int64   `json:"delta,omitempty"` // Value for counter type.
	Value *float64 `json:"value,omitempty"` // Value for gauge type.
}

// DeserializeMetrics converts bytes buffer to the slice of Metrics and returns error if anything went wrong.
func DeserializeMetrics(buf []byte) ([]Metrics, error) {

	metricsCnt := bytes.Count(buf, []byte("{"))
	switch metricsCnt {
	case 0:
		return nil, errors.New("invalid input data")
	case 1:
		var m Metrics
		if err := easyjson.Unmarshal(buf, &m); err == nil {
			return []Metrics{m}, err
		}
	default:
		metricList := make([]Metrics, 0, metricsCnt)
		if err := json.Unmarshal(buf, &metricList); err == nil {
			return metricList, nil
		}
		for idx := bytes.Index(buf, []byte{'}'}); idx >= 0 && len(buf) > idx; idx = bytes.Index(buf, []byte{'}'}) {
			var m Metrics
			if err := easyjson.Unmarshal(bytes.TrimPrefix(buf[:idx+1], []byte(",")), &m); err != nil {
				if err == io.EOF {
					break
				}
				return metricList, nil
			}
			metricList = append(metricList, m)
			idx = bytes.Index(buf, []byte{'}'})
			buf = buf[idx+1:]
		}
		return metricList, nil
	}
	return nil, errors.New("failed to deserialize metrics")
}

// SerializeMetrics serializes Metrics to the byte slice (Are you surprised? I'm not.))).
func SerializeMetrics(m *Metrics, w writer) error {
	buf, err := easyjson.Marshal(m)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

type writer interface {
	Write([]byte) (int, error)
}
