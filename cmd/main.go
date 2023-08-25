package main

import (
	"fmt"
	"net/http"
	"github.com/andrerussowsky/chat-app/internal/handlers"
)

func main() {
	http.HandleFunc("/", handlers.ServeHome)
	http.HandleFunc("/ws", handlers.ServeWebSocket)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/chat", handlers.ServeChat)

	go handlers.HandleMessages() // Start handling and broadcasting messages

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
