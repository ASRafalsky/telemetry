package transport

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/mailru/easyjson"
)

//go:generate easyjson --all json.go
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func DeserializeMetrics(buf []byte) ([]Metrics, error) {
	var metricList []Metrics
	for idx := bytes.Index(buf, []byte{'}'}); idx >= 0 && len(buf) > idx; idx = bytes.Index(buf, []byte{'}'}) {
		m := Metrics{}
		if err := easyjson.Unmarshal(buf[:idx+1], &m); err != nil {
			if err == io.EOF {
				break
			}
			return metricList, err
		}
		metricList = append(metricList, m)
		idx = bytes.Index(buf, []byte{'}'})
		buf = buf[idx+1:]
	}
	return metricList, nil
}

func SerializeMetrics(m Metrics, w writer) error {
	buf, err := json.Marshal(m)
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
