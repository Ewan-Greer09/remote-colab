package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	"github.com/Ewan-Greer09/remote-colab/views/login"
)

func HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	err := login.Page().Render(r.Context(), w)
	if err != nil {
		login.Error("Could not load page.").Render(r.Context(), w)
	}
}

func HandleLoginContent(w http.ResponseWriter, r *http.Request) {
	err := login.Content().Render(r.Context(), w)
	if err != nil {
		login.Error("Could not get content.").Render(r.Context(), w)
	}
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email-address")
	password := r.URL.Query().Get("password")

	u, err := h.DB.GetUser(email)
	if err != nil {
		err := login.NoUserWithEmail().Render(r.Context(), w)
		if err != nil {
			slog.Error("Could not render partial:", "err", err)
		}
		return
	}

	if u.Password != password {
		render.HTML(w, r, "<p>Passwords do not match</p>")
		return
	}

	cookie := http.Cookie{
		Name:     "colab-auth",
		Value:    "placeholder", //TODO: needs to be a JWT holding some form of auth token to be decoded by middleware
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}

	http.SetCookie(w, &cookie)
	w.Header().Add("HX-Location", "/")
}
