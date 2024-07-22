package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/Ewan-Greer09/remote-colab/internal/db"
	"github.com/Ewan-Greer09/remote-colab/views/chat"
)

type Room struct {
	id         string
	name       string
	db         *db.Database
	clients    map[*websocket.Conn]bool
	broadcast  chan db.Message
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
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

var (
	roomStore = make(map[string]*Room)
	roomsMu   sync.Mutex
	clientsMu sync.Mutex

	upgrade = websocket.Upgrader{
		ReadBufferSize:  32,
		WriteBufferSize: 32,
	}
)

type ChatRoom interface {
	run()
}

type WsHandler interface {
	HandleConnections(*Room, http.ResponseWriter, *http.Request)
}

func (h Handler) HandleRoomConnection(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "uid")

	roomsMu.Lock()
	room, exists := roomStore[roomID]
	roomsMu.Unlock()

	if !exists {
		rooms, err := h.DB.GetAllRooms()
		if err != nil {
			slog.Error("Could not get rooms", "err", err)
			return
		}

		for _, r := range rooms {
			if r.UID == roomID {
				room = newRoom(r.UID, r.Name, h.DB)

				roomsMu.Lock()
				roomStore[roomID] = room
				roomsMu.Unlock()

				go room.run()
				break
			}
		}

		if room == nil {
			slog.Error("Room not found", "roomID", roomID)
			return
		}
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

		room.broadcast <- *db.NewMessage(
			mc.ChatMessage,
			mc.Username,
			room.id,
			room.name,
		)
	}
}

func (room *Room) run() {
	for {
		select {
		case conn := <-room.register:
			slog.Info("Client connected", "roomID", room.id, "RoomName", room.name)

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
			if message.ChatRoom.UID == room.id && message.Content != "" {
				var buf bytes.Buffer
				err := chat.Message(
					message.Author,
					message.Author,
					message.Content,
				).Render(context.Background(), &buf)
				if err != nil {
					slog.Error("Broadcast", "err", err)
				}

				clientsMu.Lock()
				cl := room.clients
				slog.Info("Clients in room", "room", room.id, "clients", len(cl))
				for conn := range cl {
					err := conn.WriteMessage(websocket.TextMessage, buf.Bytes())
					if err != nil {
						slog.Info("Closing conn", "err", err)
						conn.Close()
						delete(room.clients, conn)
					} else {
						slog.Info("Message sent", "roomID", room.id, "author", message.Author, "roomName", room.name)
					}
				}
				clientsMu.Unlock()

				err = room.db.CreateMessage(room.id, message.Author, message.Content)
				if err != nil {
					slog.Error("Could not create message", "err", err)
				}
			}
		}
	}
}

func newRoom(roomID string, roomName string, conn *db.Database) *Room {
	return &Room{
		id:         roomID,
		name:       roomName,
		db:         conn,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan db.Message),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}
