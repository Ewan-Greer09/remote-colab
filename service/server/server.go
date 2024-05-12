package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/Ewan-Greer09/remote-colab/service/handlers"
)

type Server struct {
	*http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	if len(addr) < 4 {
		panic("addr is incorrect length")
	}

	return &Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

func NewHandler() http.HandlerFunc {
	r := chi.NewMux()
	r.Use(cors.AllowAll().Handler)

	r.Get("/", handlers.HandleRoot)
	r.Get("/index/content", handlers.HandleRootContent)

	// Serve static files
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("./service/public"))))

	return r.ServeHTTP
}
