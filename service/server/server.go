package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Ewan-Greer09/remote-colab/service/db"
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
	r.Use(
		cors.AllowAll().Handler,
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
	)

	h := handlers.Handler{
		DB: db.NewDatabase(),
	}

	r.Get("/", handlers.HandleRoot)
	r.Get("/index/content", handlers.HandleRootContent)

	r.Get("/teams", handlers.HandleTeamsPage)
	r.Get("/teams/content", handlers.HandleTeamsContent)

	r.Get("/login", handlers.HandleLoginPage)
	r.Get("/login/content", handlers.HandleLoginContent)
	r.Post("/login/submit", h.HandleUserLogin)

	// Serve static files
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("./service/public"))))

	return r.ServeHTTP
}
