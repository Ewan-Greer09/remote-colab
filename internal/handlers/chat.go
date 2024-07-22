package handlers

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/Ewan-Greer09/remote-colab/internal/db"
	m "github.com/Ewan-Greer09/remote-colab/internal/middleware"
	"github.com/Ewan-Greer09/remote-colab/views/chat"
)

func (h Handler) NewChatPage(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.DB.GetChatRoomsByUser(r.Context().Value(m.UsernameKey).(string))
	if err != nil {
		slog.Error("could not get rooms", "err", err)
		return
	}
	
	var r db.Room
	if len(room) < 1 {
		r = Room{}
	} else {
		r = rooms[0].UID
	}
	
	messages, err := h.DB.GetMessagesByRoomUID(r)
	if err != nil {
		slog.Error("could not get messages from room", "err", err, "room", rooms[0].UID)
	}
	_ = chat.NewChatPage(rooms,
		r.Context().Value(m.UsernameKey).(string),
		rooms[0].Name, messages).Render(r.Context(), w)
}

func (h Handler) ChatPage(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(m.UsernameKey).(string)
	if !ok {
		slog.Error("Could not convert to string", "key", m.UsernameKey)
		return
	}

	u, err := h.DB.GetUser(email)
	if err != nil {
		slog.Error("Could not get user", "err", err)
		return
	}

	err = chat.ChatPage("TeamWork - Chat", u.Email, true).Render(r.Context(), w)
	if err != nil {
		log.Print(err)
		return
	}
}

// func (h Handler) AvailableRooms(w http.ResponseWriter, r *http.Request) {
// 	user := chi.URLParam(r, "username")
// 	rooms, err := h.DB.GetChatRoomsByUser(user)
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		slog.Info("Could not get rooms for user", "err", err)
// 	}

// 	slog.Info("Available Rooms", "Rooms", rooms)

// 	_ = chat.AvailableRooms(rooms).Render(r.Context(), w)
// }

// func (h Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
// 	roomName := r.URL.Query().Get("room-name")
// 	email := r.Context().Value(m.UsernameKey).(string)

// 	err := h.DB.CreateRoom(db.ChatRoom{
// 		UID:      uuid.NewString(),
// 		Name:     roomName,
// 		Members:  []db.User{},
// 		Messages: []db.Message{},
// 	}, email)
// 	if err != nil {
// 		slog.Error("Could not create room", "err", err)
// 		return
// 	}

// 	rooms, err := h.DB.GetChatRoomsByUser(email)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}

// 	_ = chat.AvailableRooms(rooms).Render(r.Context(), w)
// }

func (h Handler) ChatRoom(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "uid")
	messages, err := h.DB.GetMessagesByRoomUID(roomId)
	if err != nil {
		slog.Error("could not get messages for room", "err", err)
	}

	rooms, err := h.DB.GetChatRoomsByUser(r.Context().Value(m.UsernameKey).(string))
	if err != nil {
		slog.Error("could not get rooms for user", "err", err, "user", r.Context().Value(m.UsernameKey).(string))
		render.HTML(w, r, "could not render room")
		return
	}
	var rm db.ChatRoom
	for _, room := range rooms {
		if room.UID == roomId {
			rm = room
		}
	}

	_ = chat.ChatArea(rm.Name, roomId, r.Context().Value(m.UsernameKey).(string), messages).Render(r.Context(), w)
}

func (h Handler) Invite(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("invitee")
	roomId := r.URL.Query().Get("roomId")
	u, err := h.DB.GetUser(email)
	if err != nil {
		slog.Error("Could not get user", "err", err)
		_ = chat.InviteModalVisible(roomId).Render(r.Context(), w)
		return
	}

	err = h.DB.AddUserToRoom(*u, roomId)
	if err != nil {
		slog.Error("Could not update room", "err", err)
	}

	render.HTML(w, r, "") //perhaps this should be a message to confirm success?
}

func (h Handler) InviteModal(w http.ResponseWriter, r *http.Request) {
	roomId := r.URL.Query().Get("roomId")
	_ = chat.InviteModalVisible(roomId).Render(r.Context(), w)
}

func (h Handler) InviteModalHide(w http.ResponseWriter, r *http.Request) {
	render.HTML(w, r, "")
}
