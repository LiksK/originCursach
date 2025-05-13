package handlers

import (
	"encoding/json"
	"litelife/database"
	"litelife/models"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*models.Client]bool)
var broadcast = make(chan []byte)

func LoadMessages() ([]models.ChatMessage, error) {
	rows, err := database.DB.Query(`
		SELECT id, username, message, created_at
		FROM chat_messages
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(&msg.ID, &msg.Username, &msg.Message, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, _ := store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("INSERT INTO chat_messages (username, message) VALUES ($1, $2)",
		username, message)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	msg := map[string]string{
		"username": username,
		"message":  message,
	}
	msgJSON, _ := json.Marshal(msg)
	broadcast <- msgJSON

	http.Redirect(w, r, "/user", http.StatusSeeOther)
}

func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, _ := store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	if !ok || username != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messageID := r.FormValue("message_id")
	if messageID == "" {
		http.Error(w, "Message ID required", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("DELETE FROM chat_messages WHERE id = $1", messageID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &models.Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Cleanup = func() {
		delete(clients, client)
	}

	clients[client] = true

	go client.WritePump()
	go client.ReadPump()
}

func BroadcastMessages() {
	for {
		message := <-broadcast
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
	}
}
