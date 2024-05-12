package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/Ewan-Greer09/remote-colab/views/index"
)

type Server struct {
	*http.Server
}

type Service struct {
	Server
	// db, emailer, etc...
}

func main() {
	// ...
	s := Server{
		Server: &http.Server{
			Addr:    ":3000",
			Handler: NewRouter(),
		},
	}

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

	//Views
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := index.Page().Render(context.Background(), w)
		if err != nil {
			render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
		}
	})

	r.Get("/index/content", func(w http.ResponseWriter, r *http.Request) {
		err := index.Content().Render(context.Background(), w)
		if err != nil {
			render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
		}
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
			fmt.Printf("error returning favivon: Written (%d)", n)
			return
		}
	})

	return r.ServeHTTP
}
