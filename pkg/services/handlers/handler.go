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

type dataHandler func(repository Repository, metrics transport.Metrics) (transport.Metrics, int, error)

func JSONPostHandler(repo map[string]Repository, fn dataHandler) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()
		m := transport.Metrics{}
		if err = easyjson.Unmarshal(buf, &m); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		var status int
		m, status, err = fn(repo[m.MType], m)
		if err != nil {
			http.Error(res, err.Error(), status)
			return
		}

		resBuf, err := easyjson.Marshal(&m)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		if _, err = res.Write(resBuf); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

func GaugePostHandler(repo Repository) func(http.ResponseWriter, *http.Request) {
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

		if _, err := gaugePostDataHandler(repo, key, value); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
	}
}

func GaugeGetHandler(repo Repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := gaugeGetDataHandler(repo, key)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err = io.WriteString(res, value.String())
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}

func CounterPostHandler(repo Repository) func(http.ResponseWriter, *http.Request) {
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

		if _, err := counterPostDataHandler(repo, strings.ToLower(key), value); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
	}
}

func CounterGetHandler(repo Repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := counterGetDataHandler(repo, key)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err = io.WriteString(res, value.String())
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

func AllGetHandler(tmpl *template.Template, repos map[string]Repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if len(repos) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := tmpl.Execute(res, getKeyList(repos))
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

type Repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	Delete(k string)
}
