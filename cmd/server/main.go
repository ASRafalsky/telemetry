package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ASRafalsky/telemetry/cmd/server/handlers"
	"github.com/ASRafalsky/telemetry/internal/storage"
)

func main() {
	address := parseFlags()
	fmt.Printf("Server address: %s\n", address)

	log.Fatal(http.ListenAndServe(address, newRouter()))
}

func newRouter() http.Handler {
	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/gauge/{name}/{value}", handlers.GaugePostHandler(gaugeRepo))
			r.Post("/counter/{name}/{value}", handlers.CounterPostHandler(counterRepo))
			r.Post("/{type}/{name}/{value}", handlers.FailurePostHandler())
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/gauge/{name}", handlers.GaugeGetHandler(gaugeRepo))
			r.Get("/counter/{name}", handlers.CounterGetHandler(counterRepo))
			r.Get("/{type}/{name}", handlers.FailureGetHandler())
		})
		r.Post("/", handlers.FailurePostHandler())
		r.Get("/", handlers.AllGetHandler(prepareTemplate(), gaugeRepo, counterRepo))
	})
	return r
}

func prepareTemplate() *template.Template {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Keys</title>
</head>
<body>
    <h1>Keys:</h1>
    <ul>
        {{range .}}
        <li>{{.}}</li>
        {{end}}
    </ul>
</body>
</html>
`
	return template.Must(template.New("list").Parse(tmpl))
}
