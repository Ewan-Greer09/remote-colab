package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/Ewan-Greer09/remote-colab/internal/db"
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

var (
	roomStore = make(map[string]*Room)
	roomsMu   sync.Mutex
	clientsMu sync.Mutex
)

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

func (room *Room) run() {
	for {
		select {
		case conn := <-room.register:
			slog.Info("Client connected", "room", room.Id, "RoomName", room.Name)

			clientsMu.Lock()
			room.clients[conn] = true
			clientsMu.Unlock()

		case conn := <-room.unregister:
			clientsMu.Lock()
			if _, ok := room.clients[conn]; ok {
				delete(room.clients, conn)
				conn.Close()
			}
			clientsMu.Unlock()

		case message := <-room.broadcast:
			if message.ChatRoom.UID == room.Id && message.Content != "" {
				var buf bytes.Buffer
				_ = chat.Message(chat.MessageProps{
					Content:  message.Content,
					Username: message.Author,
					Time:     time.Now(),
				}).Render(context.Background(), &buf)

				clientsMu.Lock()
				cl := room.clients
				slog.Info("Clients in room", "room", room.Id, "clients", len(cl))
				for conn := range cl {
					err := conn.WriteMessage(websocket.TextMessage, buf.Bytes())
					if err != nil {
						slog.Info("Closing conn", "err", err)
						conn.Close()
						delete(room.clients, conn)
					} else {
						slog.Info("Message sent", "room", room.Id, "author", message.Author, "RoomName", room.Name)
					}
				}
				clientsMu.Unlock()

				err := room.Handler.DB.CreateMessage(room.Id, message.Author, message.Content)
				if err != nil {
					slog.Error("Could not create message", "err", err)
				}
			}
		}
	}
}

func (h Handler) Room(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "uid")

	slog.Info("Connecting to room", "roomID", roomID)

	roomsMu.Lock()
	room, exists := roomStore[roomID]
	if !exists {
		roomsMu.Unlock()
		// Load room from database if not in memory
		rooms, err := h.DB.GetAllRooms()
		if err != nil {
			slog.Error("Could not get rooms", "err", err)
			http.Error(w, "Could not get rooms", http.StatusInternalServerError)
			return
		}

		for _, r := range rooms {
			if r.UID == roomID {
				room = &Room{
					Id:         r.UID,
					Name:       r.Name,
					Handler:    h, // for database interactions
					clients:    make(map[*websocket.Conn]bool),
					broadcast:  make(chan db.Message),
					register:   make(chan *websocket.Conn),
					unregister: make(chan *websocket.Conn),
				}
				roomsMu.Lock()
				roomStore[roomID] = room
				roomsMu.Unlock()
				go room.run()
				break
			}
		}

		if room == nil {
			slog.Error("Room not found", "roomID", roomID)
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}
	} else {
		roomsMu.Unlock()
	}

	handleConnections(room, w, r)
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
				UID:  room.Id,
				Name: room.Name,
			},
			ChatRoomID: room.Id,
		}

		room.broadcast <- m
	}
}
