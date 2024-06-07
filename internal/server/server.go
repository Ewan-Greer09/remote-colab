package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Ewan-Greer09/remote-colab/internal/db"
	"github.com/Ewan-Greer09/remote-colab/internal/handlers"
	m "github.com/Ewan-Greer09/remote-colab/internal/middleware"
)

type Server struct {
	*http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	if len(addr) < 4 {
		slog.Error("Server Address", "Length is invalid", len(addr))
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
		m.Identity,
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
		middleware.StripSlashes,
	)

	h := handlers.Handler{
		DB: db.NewDatabase(),
	}

	r.Group(func(r chi.Router) {
		r.Use(m.Auth)

		r.Get("/teams", h.HandleTeamsPage)
		r.Get("/teams/submit", h.HandleCreateTeam)
		r.Delete("/teams/{id}", h.DeleteTeam)
		r.Get("/teams/list", h.TeamsList)

		r.Get("/logout", h.Logout)

		r.Get("/chat", h.ChatPage)
		r.Get("/chat/available-rooms/{username}", h.AvailableRooms)
		r.Get("/chat/room/{uid}", h.ChatRoom)
		r.Get("/chat/room/window/{uid}", h.ChatWindow)
		r.Get("/chat/connect/{uid}", h.Room)
		r.Get("/chat/create", h.CreateRoom)
		r.Get("/chat/invite", h.Invite)
	})

	r.Get("/", handlers.HandleRoot)
	r.Get("/index/content", handlers.HandleRootContent)

	r.Get("/login", handlers.HandleLoginPage)
	r.Get("/login/content", handlers.HandleLoginContent)
	r.Get("/login/submit", h.HandleUserLogin)

	r.Get("/register", h.RegisterUserPage)
	r.Get("/register/content", handlers.RegisterUserContent)
	r.Get("/register/submit", h.RegisterUser)

	// Serve static files
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("./internal/public"))))

	return r.ServeHTTP
}
