package handlers

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

type dataHandler func(repository repository, metrics transport.Metrics) ([]byte, int, error)

func JSONPostHandler(repo repository, fn dataHandler) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()
		metricList, err := transport.DeserializeMetrics(buf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(metricList) == 0 {
			http.Error(res, "metrics list is empty", http.StatusInternalServerError)
		}
		for _, m := range metricList {
			var status int
			buf, status, err = fn(repo, m)
			if err != nil {
				http.Error(res, err.Error(), status)
				return
			}

			res.Header().Set("Content-Type", "application/json")
			if _, err = res.Write(buf); err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}

			res.WriteHeader(http.StatusOK)
		}
	}
}

func GaugePostHandler(repo repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := types.ParseGauge(chi.URLParam(req, "value"))
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		gVal := float64(value)
		if _, err := gaugePostDataHandler(repo, transport.Metrics{
			MType: Gauge,
			ID:    strings.ToLower(key),
			Value: &gVal,
		}); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
	}
}

func GaugeGetHandler(repo repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := gaugeGetDataHandler(repo, strings.ToLower(key))
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		m := transport.Metrics{}
		err = easyjson.Unmarshal(value, &m)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err = io.WriteString(res, types.Gauge(*m.Value).String())
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}

func CounterPostHandler(repo repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := types.ParseCounter(chi.URLParam(req, "value"))
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		delta := int64(value)
		if _, err := counterPostDataHandler(repo, transport.Metrics{
			MType: Counter,
			ID:    strings.ToLower(key),
			Delta: &delta,
		}); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
	}
}

func CounterGetHandler(repo repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := counterGetDataHandler(repo, strings.ToLower(key))
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		m := transport.Metrics{}
		err = easyjson.Unmarshal(value, &m)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err = io.WriteString(res, types.Counter(*m.Delta).String())
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}

func FailurePostHandler() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusBadRequest)
	}
}

func FailureGetHandler() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func AllGetHandler(tmpl *template.Template, repo repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if repo.Size() == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := tmpl.Execute(res, getKeyList(repo))
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}

func getName(req *http.Request) string {
	return chi.URLParam(req, "name")
}

type repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	Delete(k string)
}
