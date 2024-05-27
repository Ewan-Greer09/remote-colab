package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	UsernameKey    contextKey = "username"
	AuthCookieName string     = "colab-auth"
)

func Identity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var username string
		for _, cookie := range r.Cookies() {
			if cookie.Name == AuthCookieName {
				username = cookie.Value
			}
		}

		if username == "" {
			username = "Undefined"
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UsernameKey, username)))
	})
}
