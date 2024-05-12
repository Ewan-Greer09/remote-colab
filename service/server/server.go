package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

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

func NewHandler() http.Handler {
	r := chi.NewMux()

	r.Get("/", handlers.HandleRoot)
	r.Get("/index/content", handlers.HandleRootContent)

	return r
}
