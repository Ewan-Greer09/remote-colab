package handlers

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/Ewan-Greer09/remote-colab/service/db"
	m "github.com/Ewan-Greer09/remote-colab/service/middleware"
	"github.com/Ewan-Greer09/remote-colab/views/chat"
)

func (h Handler) ChatPage(w http.ResponseWriter, r *http.Request) {
	email := r.Context().Value(m.UsernameKey).(string)

	u, err := h.DB.GetUser(email)
	if err != nil {
		slog.Error("oops")
		return
	}

	err = chat.ChatPage(u.Email, u.DisplayName).Render(r.Context(), w)
	if err != nil {
		log.Print(err)
		return
	}
}

func (h Handler) AvailableRooms(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "username")
	rooms, err := h.DB.GetChatRoomsByUser(user)
	if err != nil {
		slog.Info("Could not get rooms for user", "err", err)
	}

	_ = chat.AvailableRooms(rooms).Render(r.Context(), w)
}

func (h Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Query().Get("room-name")
	email := r.Context().Value(m.UsernameKey).(string)

	err := h.DB.CreateRoom(db.ChatRoom{
		UID:      uuid.NewString(),
		Name:     roomName,
		Members:  []db.User{},
		Messages: []db.Message{},
	}, email)
	if err != nil {
		slog.Error("Could not create room", "err", err)
	}

	rooms, err := h.DB.GetChatRoomsByUser(email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_ = chat.AvailableRooms(rooms).Render(r.Context(), w)
}

func (h Handler) ChatRoom(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "uid")
	username := r.Context().Value(m.UsernameKey).(string)

	_ = chat.ChatRoom(username, roomId).Render(r.Context(), w)
}

func (h Handler) ChatWindow(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "uid")
	messages, err := h.DB.GetMessagesByRoomUID(roomId)
	if err != nil {
		slog.Error("Could not get messages", "err", err)
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	u, err := h.DB.GetUser(r.Context().Value(m.UsernameKey).(string))
	if err != nil {
		slog.Error("Could not get user", "err", err)
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	err = chat.ChatWindow(chat.ChatWindowProps{Username: u.DisplayName, RoomID: roomId, Messages: messages}).Render(r.Context(), w)
	if err != nil {
		slog.Error("Could not render ChatWindow", "err", err)
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}
}

func (h Handler) Invite(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("invitee")
	roomId := r.URL.Query().Get("roomId")
	u, err := h.DB.GetUser(email)
	if err != nil {
		slog.Error("Could not get user", "err", err)
		return
	}

	err = h.DB.AddUserToRoom(*u, roomId)
	if err != nil {
		slog.Error("Could not update room", "err", err)
	}
}
