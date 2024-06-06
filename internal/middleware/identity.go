package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	UsernameKey     contextKey = "username"
	AuthCookieName  string     = "colab-auth"
	AuthedKey       contextKey = "authed"
	DefaultUsername string     = "Undefined"
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
			username = DefaultUsername
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UsernameKey, username)))
	})
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.Context().Value(UsernameKey).(string)) < 3 || r.Context().Value(UsernameKey).(string) == DefaultUsername {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AuthedKey, true)))
	})
}
