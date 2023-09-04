package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"

	"github.com/andrerussowsky/chat-app/internal/models"
)

var (
	clients         = make(map[*websocket.Conn]bool)
	broadcast       = make(chan models.Message)
	messages        = []models.Message{}
	stockQuotes     = make(chan models.Message)
	maxMessageCount = 50
)

// upgrader is used to upgrade the HTTP connection to a WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// ServeWebSocket handles WebSocket requests from the peer
func ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Handle error
		return
	}
	defer conn.Close()

	// Add the new connection to the clients map
	clients[conn] = true

	for {
		var message models.Message
		err := conn.ReadJSON(&message)

		if err != nil {
			// Handle error and remove connection from clients map
			delete(clients, conn)
			return
		}

		username, err := ParseJWTToken(message.Token)
		if err != nil {
			delete(clients, conn)
			return
		}

		if strings.HasPrefix(message.Content, "/") {
			if strings.HasPrefix(message.Content, "/stock=") {
				stockCode := strings.TrimPrefix(message.Content, "/stock=")
				// Call the bot's API with the stock code
				go callBotAPI(stockCode)
				continue
			}

			sendBotMessage("")
			continue
		}

		message.Username = username
		message.Timestamp = time.Now().Format(time.DateTime)

		// Send the received message to the broadcast channel
		broadcast <- message
	}
}

// HandleMessages handles the broadcast channel
func HandleMessages() {
	for {
		// Retrieve the next message from the broadcast channel
		message := <-broadcast

		// When adding a new message:
		if len(messages) >= maxMessageCount {
			// Remove the oldest message
			messages = messages[1:]
		}

		messages = append(messages, message)

		// Broadcast the message to all connected clients
		for client := range clients {
			err := client.WriteJSON(messages)
			if err != nil {
				// Handle error and remove connection from clients map
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// ServeChat handles HTTP requests for the chat page
func ServeChat(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token := r.URL.Query().Get("token")

		username, err := ParseJWTToken(token)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check if user is authenticated, otherwise redirect to login
		session, err := store.Get(r, username)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if r.Method == http.MethodGet {
			// Create a data struct to pass to the template
			data := struct {
				Token    string
				Messages []models.Message
			}{
				Token:    token,
				Messages: messages,
			}

			// Serve the chat page
			templates.ExecuteTemplate(w, "chat.html", data)
		}
	}
}

func callBotAPI(stockCode string) {
	url := fmt.Sprintf("http://localhost:8082/stock-quote?stock_code=%s", stockCode)
	_, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to declare a queue: %v", err)
	}
}

func ConsumeStockQuotes() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://chat-app-user:chat-app-password@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"stock_quotes",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	for msg := range msgs {
		sendBotMessage(string(msg.Body))
	}
}

func sendBotMessage(message string) {
	if message == "" {
		message = "I'm sorry, I didn't understand that command. Please use /stock=stock_code format to get stock quotes."
	}

	stockQuote := models.Message{
		Username:  "Bot",
		Content:   message,
		Timestamp: time.Now().Format(time.DateTime),
	}
	broadcast <- stockQuote
}
