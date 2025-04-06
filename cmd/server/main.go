package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

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
		r.Get("/", allGetHandler(prepareTemplate(), gaugeRepo, counterRepo))
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
