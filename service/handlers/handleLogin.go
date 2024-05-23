package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	"github.com/Ewan-Greer09/remote-colab/views/login"
)

func HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	err := login.Page().Render(r.Context(), w)
	if err != nil {
		render.HTML(w, r, "<p>There was an issue</p>")
	}
}

func HandleLoginContent(w http.ResponseWriter, r *http.Request) {
	err := login.Content(login.LoginData{
		Text: "This is some placeholder text.",
	}).Render(r.Context(), w)
	if err != nil {
		render.HTML(w, r, "<p>There was an issue</p>")
	}
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email-address")
	password := r.URL.Query().Get("password")

	u, err := h.DB.GetUser(email)
	if err != nil {
		slog.Info("Could not get user:", "err", err)
		err := login.NoUserWithEmail().Render(r.Context(), w)
		if err != nil {
			slog.Error("Could not render partial:", "err", err)
		}
		return
	}

	if u.Password != password {
		slog.Info("Passwords do not match")
		render.HTML(w, r, "<p>Passwords do not match</p>")
		return
	}

	cookie := &http.Cookie{
		Name:     "colab-auth",
		Value:    "placeholder", //TODO: needs to be a JWT holding some form of auth token to be decoded by middleware
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}

	http.SetCookie(w, cookie)
	render.HTML(w, r, fmt.Sprintf("%s %s", email, password))
}
