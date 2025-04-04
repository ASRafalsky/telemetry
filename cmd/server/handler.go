package main

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func gaugePostHandler(repo GaugeRepository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		key := getName(req)
		if len(key) == 0 {
			res.WriteHeader(http.StatusNotFound)
		}

		floatValue, err := strconv.ParseFloat(chi.URLParam(req, "value"), 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}

		repo.Set(strings.ToLower(key), floatValue)
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
			valStr := strconv.FormatFloat(val, 'g', -1, 64)
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, err := io.WriteString(res, valStr)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
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

		intValue, err := strconv.ParseInt(chi.URLParam(req, "value"), 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		repo.Set(strings.ToLower(key), intValue)
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
			valStr := strconv.FormatInt(val, 10)
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, err := io.WriteString(res, valStr)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
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

func allGetHandler(repos []CommonRepository) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var result []string
		for _, repo := range repos {
			result = append(result, repo.Keys()...)
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := res.Write([]byte(strings.Join(result, ","))); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
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
	Set(k string, v float64)
	Get(k string) (float64, bool)
	ForEach(ctx context.Context, fn func(k string, v float64) error) error
}

type CounterRepository interface {
	CommonRepository
	Set(k string, v int64)
	Get(k string) (int64, bool)
	ForEach(ctx context.Context, fn func(k string, v int64) error) error
}
