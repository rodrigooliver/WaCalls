package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed docs/openapi.yaml docs/index.html
var docsFS embed.FS

func mountDocs(mux *http.ServeMux) {
	sub, err := fs.Sub(docsFS, "docs")
	if err != nil {
		panic(err)
	}
	docs := http.FileServer(http.FS(sub))
	mux.Handle("GET /docs/", http.StripPrefix("/docs/", docs))
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusFound)
	})
}
