package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/Ewan-Greer09/remote-colab/internal/db"
	m "github.com/Ewan-Greer09/remote-colab/internal/middleware"
	"github.com/Ewan-Greer09/remote-colab/views/teams"
)

func (h Handler) HandleTeamsPage(w http.ResponseWriter, r *http.Request) {
	t, err := h.DB.GetTeamsForUser(r.Context().Value(m.UsernameKey).(string))
	if err != nil {
		slog.Info("Could not get teams for user", "err", err)
	}
	err = teams.Page("TeamWork - Teams", true, t).Render(context.Background(), w)
	if err != nil {
		render.JSON(w, r, fmt.Errorf("there was an issue: %w", err))
	}
}

func (h Handler) HandleCreateTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.FormValue("team-name")
	teamDesc := r.FormValue("team-description")

	u, err := h.DB.GetUser(r.Context().Value(m.UsernameKey).(string))
	if err != nil {
		slog.Error("Could not get user", "err", err)
		return
	}

	err = h.DB.CreateTeamForUser(db.Team{
		UID:         uuid.NewString(),
		Name:        teamName,
		Description: teamDesc,
		Members:     []db.User{},
	}, *u)
	if err != nil {
		slog.Error("could not create team", "err", err)
		return
	}

	_ = teams.CreateTeamForm().Render(r.Context(), w)
}

func (h Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.DB.DeleteTeam(id)
	if err != nil {
		render.HTML(w, r, "could not delete team")
		return
	}

	t, err := h.DB.GetTeamsForUser(r.Context().Value(m.UsernameKey).(string))
	if err != nil {
		slog.Info("Could not get teams for user", "err", err)
	}

	_ = teams.TeamsList(t).Render(r.Context(), w)
}

func (h Handler) TeamsList(w http.ResponseWriter, r *http.Request) {
	t, err := h.DB.GetTeamsForUser(r.Context().Value(m.UsernameKey).(string))
	if err != nil {
		slog.Error("could not get teams for user", "err", err)
	}
	_ = teams.TeamsList(t).Render(r.Context(), w)
}
