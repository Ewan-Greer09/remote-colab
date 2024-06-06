package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	m "github.com/Ewan-Greer09/remote-colab/internal/middleware"
	"github.com/Ewan-Greer09/remote-colab/views/index"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	var username string
	var loggedIn bool
	for _, cookie := range r.Cookies() {
		if cookie.Name == m.AuthCookieName {
			username = cookie.Value
		}
	}

	if username != "" {
		loggedIn = true
	}

	err := index.Page("TeamWork - Home", loggedIn).Render(r.Context(), w)
	if err != nil {
		render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
	}
}

func HandleRootContent(w http.ResponseWriter, r *http.Request) {
	// data := index.IndexData{}
	// data.IntroText = "This is some intro text for the index page. This will eventually be a team management tool."

	// err := index.Content(data).Render(context.Background(), w)
	// if err != nil {
	// 	render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
	// }
}
