package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/Ewan-Greer09/remote-colab/views/index"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	err := index.Page().Render(context.Background(), w)
	if err != nil {
		render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
	}
}

func HandleRootContent(w http.ResponseWriter, r *http.Request) {
	data := index.IndexData{}
	data.IntroText = "This is some intro text for the index page. This will eventually be a team management tool."

	err := index.Content(data).Render(context.Background(), w)
	if err != nil {
		render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
	}
}
