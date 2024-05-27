package handlers

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
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

	log.Printf("Display Name: %s", u.DisplayName)

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
		UID:      uuid.New(),
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
	var username string
	for _, cookie := range r.Cookies() {
		if cookie.Name == m.AuthCookieName {
			username = cookie.Value
		}
	}

	_ = chat.ChatRoom(username, roomId).Render(r.Context(), w)
}

func (h Handler) ChatWindow(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "uid")

	u, err := h.DB.GetUser(r.Context().Value(m.UsernameKey).(string))
	if err != nil {
		slog.Error("Could not get user", "err", err)
	}
	_ = chat.ChatWindow(chat.ChatWindowProps{Username: u.DisplayName, RoomID: roomId}).Render(r.Context(), w)
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
