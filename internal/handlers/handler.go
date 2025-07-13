package handlers

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

type dataHandler func(ctx context.Context, repository repository, metrics transport.Metrics) ([]byte, int, error)

func JSONPostHandler(repo repository, fn dataHandler) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		defer func() {
			// Handled at the logging level.
			_ = req.Body.Close()
		}()
		metricList, err := transport.DeserializeMetrics(buf)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(metricList) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		for _, m := range metricList {
			var status int
			buf, status, err = fn(req.Context(), repo, m)
			if err != nil {
				res.WriteHeader(status)
				return
			}

			res.Header().Set("Content-Type", "application/json")
			if _, err = res.Write(buf); err != nil {
				res.WriteHeader(http.StatusInternalServerError)
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
			MType: types.GaugeType,
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

		value, err := gaugeGetDataHandler(req.Context(), repo, strings.ToLower(key))
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
		if _, err := counterPostDataHandler(req.Context(), repo, transport.Metrics{
			MType: types.CounterType,
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

		value, err := counterGetDataHandler(req.Context(), repo, strings.ToLower(key))
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

func DBPingHandler(repo repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctxPing, cancel := context.WithTimeout(req.Context(), time.Second)
		defer cancel()
		if err := repo.Ping(ctxPing); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}

func AllGetHandler(tmpl *template.Template, repo repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		keys, err := getKeyList(repo)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(res, keys)
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
	Get(ctx context.Context, k string) ([]byte, error)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Delete(ctx context.Context, k string) error
	Ping(ctx context.Context) error
	Size() (int, error)
}
