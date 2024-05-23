package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"gorm.io/gorm"

	"github.com/Ewan-Greer09/remote-colab/service/db"
	"github.com/Ewan-Greer09/remote-colab/views/register"
)

func (h *Handler) RegisterUserPage(w http.ResponseWriter, r *http.Request) {
	err := register.Page().Render(r.Context(), w)
	if err != nil {
		render.HTML(w, r, "<p>There was an issue</p>")
	}
}

func RegisterUserContent(w http.ResponseWriter, r *http.Request) {
	err := register.Content().Render(r.Context(), w)
	if err != nil {
		render.HTML(w, r, "<p id='error'>There was an issue</p>")
	}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email-address")
	password := r.URL.Query().Get("password")

	_, err := h.DB.GetUser(email)
	if err != gorm.ErrRecordNotFound {
		err := register.Error("Email already in use.").Render(r.Context(), w)
		if err != nil {
			slog.Error("Could not return error message", "err", err)
		}
		return
	}

	user := db.User{
		Email:    email,
		Password: password,
	}

	err = h.DB.CreateUser(user)
	if err != nil {
		slog.Error("Could not create user", "err", err)
		err := register.Error("Could not create user").Render(r.Context(), w)
		if err != nil {
			slog.Error("Could not return error message", "err", err)
		}
		return
	}

	w.Header().Add("HX-Location", "/login")
	render.HTML(w, r, "<p id='error'>User created successfully</p>")
}
