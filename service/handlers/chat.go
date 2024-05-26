package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

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
	var email string
	roomName := r.URL.Query().Get("room-name")
	for _, cookie := range r.Cookies() {
		if cookie.Name == m.AuthCookieName {
			email = cookie.Value
		}
	}

	err := h.DB.CreateRoom(email, roomName)
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

// ws - section
var upgrade = websocket.Upgrader{
	ReadBufferSize:  32,
	WriteBufferSize: 32,
}

type Room struct {
	Id         string
	Name       string
	clients    map[*websocket.Conn]bool
	broadcast  chan ChatMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

type ChatMessage struct {
	Content MessageContent
	Sender  string
	RoomId  string
	SentAt  time.Time
}

var rooms = make(map[string]*Room)
var mu sync.Mutex

func newRoom(name string) *Room {
	return &Room{
		Id:         uuid.NewString(),
		Name:       name,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan ChatMessage),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (room *Room) run() {
	for {
		select {
		case conn := <-room.register:
			room.clients[conn] = true
		case conn := <-room.unregister:
			if _, ok := room.clients[conn]; ok {
				delete(room.clients, conn)
				conn.Close()
			}
		case message := <-room.broadcast:
			if message.RoomId == room.Id {
				for conn := range room.clients {
					var buf bytes.Buffer

					err := chat.Message(
						chat.MessageProps{
							Content:  message.Content.ChatMessage,
							Username: message.Sender,
							Time:     time.Now(),
						},
					).Render(context.Background(), &buf)
					if err != nil {
						conn.Close()
						delete(room.clients, conn)
					}

					err = conn.WriteMessage(websocket.TextMessage, buf.Bytes())
					if err != nil {
						conn.Close()
						delete(room.clients, conn)
					}
				}
			}
		}
	}
}

func (h Handler) Room(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "uid")

	mu.Lock()
	if _, ok := rooms[roomID]; !ok {
		rooms[roomID] = newRoom("")
		go rooms[roomID].run()
		slog.Info("Running new room", "roomId", roomID)
	}
	mu.Unlock()
	handleConnections(rooms[roomID], w, r)
}

type MessageContent struct {
	ChatMessage string `json:"chat-message"`
	Username    string `json:"username"`
	Headers     struct {
		HXRequest     string `json:"HX-Request"`
		HXTrigger     string `json:"HX-Trigger"`
		HXTriggerName any    `json:"HX-Trigger-Name"`
		HXTarget      string `json:"HX-Target"`
		HXCurrentURL  string `json:"HX-Current-URL"`
	} `json:"HEADERS"`
}

func handleConnections(room *Room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Could not upgrade connection", "err", err)
		return
	}
	defer conn.Close()
	room.register <- conn

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			room.unregister <- conn
			break
		}

		var mc MessageContent
		err = json.Unmarshal(message, &mc)
		if err != nil {
			return
		}

		m := ChatMessage{
			Content: mc,
			Sender:  mc.Username,
			RoomId:  room.Id,
			SentAt:  time.Now().UTC(),
		}
		room.broadcast <- m
	}
}
