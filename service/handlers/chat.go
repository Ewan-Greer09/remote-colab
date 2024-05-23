package handlers

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/Ewan-Greer09/remote-colab/views/chat"
)

func (h *Handler) ChatPage(w http.ResponseWriter, r *http.Request) {
	err := chat.Page().Render(r.Context(), w)
	if err != nil {
		slog.Error("Could not render page", "err", err)
		return
	}
}

func (h *Handler) ChatContent(w http.ResponseWriter, r *http.Request) {
	cookies := r.Cookies()
	var username string
	for cookie := range cookies {
		if cookies[cookie].Name == "colab-auth" {
			username = cookies[cookie].Value
		}
	}

	_ = chat.Content(username).Render(r.Context(), w)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var clientMu sync.RWMutex

type Message struct {
	Username string `json:"username"`
	Msg      string `json:"chat-message"`
}

func (h *Handler) ChatWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Upgrader", "err", err)
		return
	}
	defer conn.Close()

	clientMu.Lock()
	clients[conn] = true
	clientMu.Unlock()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			slog.Error("Read JSON", "err", err)
			delete(clients, conn)
			return
		}

		broadcast <- msg
	}
}

func HandleMessages() {
	for {
		for client := range clients {
			log.Printf("%+v", client.LocalAddr())
		}
		msg := <-broadcast

		for client := range clients {
			var buf = bytes.Buffer{}
			err := chat.Message(msg.Msg, msg.Username).Render(context.Background(), &buf)
			if err != nil {
				slog.Error("Error rendering message", "err", err)
				continue
			}
			err = client.WriteMessage(1, buf.Bytes())
			if err != nil {
				slog.Error("Write JSON", "err", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
