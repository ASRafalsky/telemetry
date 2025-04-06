package main

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ASRafalsky/telemetry/internal/types"
)

func gaugePostHandler(repo Repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
		}

		value, err := types.ParseGauge(chi.URLParam(req, "value"))
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}

		repo.Set(strings.ToLower(key), types.GaugeToBytes(value))
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
}

func gaugeGetHandler(repo Repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		if value, ok := repo.Get(strings.ToLower(key)); ok {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, err := io.WriteString(res, types.BytesToGauge(value).String())
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		res.WriteHeader(http.StatusNotFound)
	}
}

func counterPostHandler(repo Repository) func(http.ResponseWriter, *http.Request) {
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

		if previousValue, ok := repo.Get(strings.ToLower(key)); ok {
			value += types.BytesToCounter(previousValue)
		}
		repo.Set(strings.ToLower(key), types.CounterToBytes(value))
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
}

func counterGetHandler(repo Repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		if value, ok := repo.Get(strings.ToLower(key)); ok {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, err := io.WriteString(res, types.BytesToCounter(value).String())
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		res.WriteHeader(http.StatusNotFound)
	}
}

func failurePostHandler() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
		}
		res.WriteHeader(http.StatusBadRequest)
	}
}

func failureGetHandler() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func allGetHandler(tmpl *template.Template, repos ...Repository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if len(repos) == 0 {
			res.WriteHeader(http.StatusNotFound)
		}

		totalEntryCnt := 0
		for _, repo := range repos {
			totalEntryCnt += repo.Size()
		}

		result := make([]string, totalEntryCnt)
		for _, repo := range repos {
			result = append(result, repo.Keys()...)
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := tmpl.Execute(res, result)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getName(req *http.Request) string {
	return chi.URLParam(req, "name")
}

type Repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Keys() []string
	Size() int
	Delete(k string)
}
