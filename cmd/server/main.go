package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ASRafalsky/telemetry/internal/storage"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", newRouter()))
}

func newRouter() http.Handler {
	gaugeRepo := storage.New[string, float64]()
	counterRepo := storage.New[string, int64]()

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/gauge/{name}/{value}", gaugePostHandler(gaugeRepo))
			r.Post("/counter/{name}/{value}", counterPostHandler(counterRepo))
			r.Post("/{type}/{name}/{value}", failurePostHandler())
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/gauge/{name}", gaugeGetHandler(gaugeRepo))
			r.Get("/counter/{name}", counterGetHandler(counterRepo))
			r.Get("/{type}/{name}", failureGetHandler())
		})
		r.Post("/", failurePostHandler())
		r.Get("/", allGetHandler([]CommonRepository{gaugeRepo, counterRepo}))
	})
	return r
}
