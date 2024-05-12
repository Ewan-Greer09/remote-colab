package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/a-h/templ" // needed to prevent "go mod tidy" from breaking templ functions
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Ewan-Greer09/remote-colab/service/server"
)

type Service struct {
	server.Server
	// db, emailer, etc...
}

func main() {
	s := server.NewServer(":3000", server.NewHandler())
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func NewRouter() http.HandlerFunc {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	//Data routes
	r.Route("/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World - V1 API"))
		})
	})

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("public/img/collab-logo.png")
		if err != nil {
			fmt.Printf("There was an error, %s", err)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "Image/png")

		n, err := io.Copy(w, f)
		if err != nil {
			fmt.Printf("error returning favivon: Written (%d) bytes", n)
			return
		}
	})

	return r.ServeHTTP
}
