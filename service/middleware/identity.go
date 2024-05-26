package middleware

import (
	"context"
	"net/http"

	"github.com/Ewan-Greer09/remote-colab/service/handlers"
)

type contextKey string

const (
	usernameKey contextKey = "username"
)

func Identity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var username string
		for _, cookie := range r.Cookies() {
			if cookie.Name == handlers.AuthCookieName {
				username = cookie.Value
			}
		}

		if username == "" {
			next.ServeHTTP(w, r)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), usernameKey, username))
		next.ServeHTTP(w, r)
	})
}
