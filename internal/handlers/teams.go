package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/Ewan-Greer09/remote-colab/views/teams"
)

func HandleTeamsPage(w http.ResponseWriter, r *http.Request) {
	err := teams.Page("TeamWork - Teams", true).Render(context.Background(), w)
	if err != nil {
		render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
	}
}

func HandleTeamsContent(w http.ResponseWriter, r *http.Request) {
	data := teams.TeamsData{
		Text: "This is some placeholder text.",
	}

	err := teams.Content(data).Render(context.Background(), w)
	if err != nil {
		render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
	}
}
