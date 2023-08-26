package main

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/andrerussowsky/chat-app/internal/handlers"
)

var templates *template.Template

func main() {
	templates := loadTemplates() // Load templates

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static")))) // Serve static files

	http.HandleFunc("/", handlers.ServeHome(templates))               // Serve index
	http.HandleFunc("/register", handlers.RegisterHandler(templates)) // Register user
	http.HandleFunc("/login", handlers.LoginHandler(templates))       // Login user
	http.HandleFunc("/chat", handlers.ServeChat(templates))           // Serve chat

	http.HandleFunc("/ws", handlers.ServeWebSocket) // Serve websocket

	go handlers.HandleMessages()     // Start handling and broadcasting messages
	go handlers.ConsumeStockQuotes() // Start consuming stock quotes

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil) // Start server
}

func loadTemplates() *template.Template {
	return template.Must(template.ParseFiles("templates/register.html", "templates/login.html", "templates/chat.html", "static/index.html"))
}
