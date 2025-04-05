package main

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ASRafalsky/telemetry/internal"
)

func gaugePostHandler(repo GaugeRepository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
		}

		value, err := internal.ParseGauge(chi.URLParam(req, "value"))
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}

		repo.Set(strings.ToLower(key), value)
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
}

func gaugeGetHandler(repo GaugeRepository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		if val, ok := repo.Get(strings.ToLower(key)); ok {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, err := io.WriteString(res, val.String())
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		res.WriteHeader(http.StatusNotFound)
	}
}

func counterPostHandler(repo CounterRepository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := internal.ParseCounter(chi.URLParam(req, "value"))
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		if previousValue, ok := repo.Get(strings.ToLower(key)); ok {
			value += previousValue
		}
		repo.Set(strings.ToLower(key), value)
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
}

func counterGetHandler(repo CounterRepository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		if val, ok := repo.Get(strings.ToLower(key)); ok {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, err := io.WriteString(res, val.String())
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

func allGetHandler(tmpl *template.Template, repos ...CommonRepository) func(http.ResponseWriter, *http.Request) {
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

type CommonRepository interface {
	Keys() []string
	Size() int
	Delete(k string)
}

type GaugeRepository interface {
	CommonRepository
	Set(k string, v internal.Gauge)
	Get(k string) (internal.Gauge, bool)
	ForEach(ctx context.Context, fn func(k string, v internal.Gauge) error) error
}

type CounterRepository interface {
	CommonRepository
	Set(k string, v internal.Counter)
	Get(k string) (internal.Counter, bool)
	ForEach(ctx context.Context, fn func(k string, v internal.Counter) error) error
}
