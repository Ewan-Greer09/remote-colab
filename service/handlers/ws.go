package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/Ewan-Greer09/remote-colab/service/db"
	m "github.com/Ewan-Greer09/remote-colab/service/middleware"
	"github.com/Ewan-Greer09/remote-colab/views/chat"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  32,
	WriteBufferSize: 32,
}

type Room struct {
	Id         string
	Name       string
	Handler    Handler
	clients    map[*websocket.Conn]bool
	broadcast  chan db.Message
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// MessageContent is the struct that will be used to unmarshal the message sent by the client - without this the message is a mess
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

func newRoom(name string, email string) *Room {
	roomId := uuid.New()
	room := &Room{
		Id:   roomId.String(),
		Name: name,
		Handler: Handler{
			DB: db.NewDatabase(),
		},
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan db.Message),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}

	room.Handler.DB.CreateRoom(db.ChatRoom{
		UID:      roomId,
		Name:     name,
		Members:  []db.User{},
		Messages: []db.Message{},
	}, email)
	return room
}

func (room *Room) run() {
	slog.Info("Running room", "room", room.Id, "name", room.Name)
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
			if message.ChatRoom.UID.String() == room.Id {
				for conn := range room.clients {
					var buf bytes.Buffer
					err := chat.Message(chat.MessageProps{
						Content:  message.Content,
						Username: message.Author,
						Time:     time.Now(),
					}).Render(context.Background(), &buf)
					if err != nil {
						conn.Close()
						delete(room.clients, conn)
					}

					err = conn.WriteMessage(websocket.TextMessage, buf.Bytes())
					if err != nil {
						conn.Close()
						delete(room.clients, conn)
					}

					err = room.Handler.DB.CreateMessage(room.Id, message.Author, message.Content)
					if err != nil {
						slog.Error("Could not create message", "err", err)
						return
					}

					slog.Info("Message sent", "room", room.Id, "author", message.Author, "RoomName", room.Name)
				}
			}
		}
	}
}

func (h Handler) Room(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "uid")
	rooms, err := h.DB.GetAllRooms()
	if err != nil {
		slog.Error("Could not get rooms", "err", err)
		return
	}

	// get room by id and if gorm.ErrRecordNotFound create a new record

	var room Room
	for _, r := range rooms {
		if r.UID.String() == roomID {
			room = Room{
				Id:         r.UID.String(),
				Name:       r.Name,
				Handler:    h,
				clients:    make(map[*websocket.Conn]bool),
				broadcast:  make(chan db.Message),
				register:   make(chan *websocket.Conn),
				unregister: make(chan *websocket.Conn),
			}
			break
		}
	}

	if room.Id == "" {
		r := newRoom(roomID, r.Context().Value(m.AuthCookieName).(string))
		room = *r
	}

	go room.run()
	handleConnections(&room, w, r)
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

		m := db.Message{
			Content: mc.ChatMessage,
			Author:  mc.Username,
			ChatRoom: db.ChatRoom{
				UID:  uuid.MustParse(room.Id),
				Name: room.Name,
			},
			ChatRoomID: room.Id,
		}

		room.broadcast <- m
	}
}
