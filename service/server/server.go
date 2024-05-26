package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Ewan-Greer09/remote-colab/service/db"
	"github.com/Ewan-Greer09/remote-colab/service/handlers"
	m "github.com/Ewan-Greer09/remote-colab/service/middleware"
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
		middleware.StripSlashes,
		m.Identity,
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
	r.Get("/login/submit", h.HandleUserLogin)

	r.Get("/register", h.RegisterUserPage)
	r.Get("/register/content", handlers.RegisterUserContent)
	r.Get("/register/submit", h.RegisterUser)

	r.Get("/chat", h.ChatPage)
	r.Get("/chat/available-rooms/{username}", h.AvailableRooms)
	r.Get("/chat/room/{uid}", h.ChatRoom)
	r.Get("/chat/room/window/{uid}", h.ChatWindow)
	r.Get("/chat/connect/{uid}", h.Room)
	r.Get("/chat/create", h.CreateRoom)
	r.Get("/chat/invite", h.Invite)

	// r.Get("/chat/connect", )

	// Serve static files
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("./service/public"))))

	return r.ServeHTTP
}
