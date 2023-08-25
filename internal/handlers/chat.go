package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

type Message struct {
	Username  string `json:"username"`
	Content   string `json:"content"`
	Token     string `json:"token"`
	Timestamp string `json:"timestamp"`
}

var messages = []Message{}

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
		var message Message
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

		if strings.HasPrefix(message.Content, "/stock=") {
			stockCode := strings.TrimPrefix(message.Content, "/stock=")
			// Call the bot's API with the stock code
			go callBotAPI(stockCode)
			continue
		}

		message.Username = username
		message.Timestamp = time.Now().Format(time.DateTime)

		// Send the received message to the broadcast channel
		broadcast <- message
	}
}

func HandleMessages() {
	for {
		// Retrieve the next message from the broadcast channel
		message := <-broadcast

		messages = append(messages, message)

		// Broadcast the message to all connected clients
		for client := range clients {
			err := client.WriteJSON(message)
			if err != nil {
				// Handle error and remove connection from clients map
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func ServeChat(w http.ResponseWriter, r *http.Request) {
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
		// Retrieve the last 50 messages (or fewer if less than 50 messages exist)
		start := 0
		if len(messages) > 50 {
			start = len(messages) - 50
		}
		recentMessages := messages[start:]

		// Create a data struct to pass to the template
		data := struct {
			Token    string
			Messages []Message
		}{
			Token:    token,
			Messages: recentMessages,
		}

		// Serve the chat page
		templates.ExecuteTemplate(w, "chat.html", data)
	}
}

func callBotAPI(stockCode string) {
	url := fmt.Sprintf("http://localhost:8082/stock-quote?stock_code=%s", stockCode)
	_, err := http.Get(url)
	if err != nil {
		// Handle error
	}
}

var stockQuotes = make(chan Message)

func consumeStockQuotes() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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
		stockQuote := Message{
			Username:  "Bot",
			Content:   string(msg.Body),
			Timestamp: time.Now().Format(time.DateTime),
		}
		broadcast <- stockQuote
	}
}

func init() {
	// Start consuming stock quotes from RabbitMQ
	go consumeStockQuotes()
}
