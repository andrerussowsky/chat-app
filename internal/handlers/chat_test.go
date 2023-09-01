package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/andrerussowsky/chat-app/internal/models"
	"github.com/gorilla/websocket"
)

func TestServeWebSocket(t *testing.T) {
	mockToken := "mock-jwt-token"
	mockMessage := models.Message{
		Token:    mockToken,
		Content:  "Hello, world!",
		Username: "testuser",
	}

	// Create a mock WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		// Simulate sending a mock message
		err := conn.WriteJSON(mockMessage)
		if err != nil {
			t.Errorf("failed to write mock message to WebSocket connection: %v", err)
		}
	}))
	defer server.Close()

	// Connect to the mock WebSocket server
	u := "ws" + strings.TrimPrefix(server.URL, "http") + "/"
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// Read the message sent by the mock WebSocket server
	var receivedMessage models.Message
	err = conn.ReadJSON(&receivedMessage)
	if err != nil {
		t.Errorf("failed to read message from WebSocket connection: %v", err)
	}

	// Validate the received message
	if receivedMessage.Token != mockToken {
		t.Errorf("unexpected token in received message: got %v, want %v", receivedMessage.Token, mockToken)
	}
	if receivedMessage.Content != mockMessage.Content {
		t.Errorf("unexpected content in received message: got %v, want %v", receivedMessage.Content, mockMessage.Content)
	}
	if receivedMessage.Username != mockMessage.Username {
		t.Errorf("unexpected username in received message: got %v, want %v", receivedMessage.Username, mockMessage.Username)
	}
}

func TestHandleMessages(t *testing.T) {
	// Create a mock WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		// Create a mock message
		mockMessage := models.Message{
			Token:    "mock-token",
			Content:  "Hello, world!",
			Username: "testuser",
		}

		// Send the mock message to the broadcast channel
		broadcast <- mockMessage

		// Read the message sent by the broadcast channel
		var receivedMessage models.Message
		err := conn.ReadJSON(&receivedMessage)
		if err != nil {
			t.Errorf("failed to read message from WebSocket connection: %v", err)
		}

		// Validate the received message
		if receivedMessage.Token != mockMessage.Token {
			t.Errorf("unexpected token in received message: got %v, want %v", receivedMessage.Token, mockMessage.Token)
		}
		if receivedMessage.Content != mockMessage.Content {
			t.Errorf("unexpected content in received message: got %v, want %v", receivedMessage.Content, mockMessage.Content)
		}
		if receivedMessage.Username != mockMessage.Username {
			t.Errorf("unexpected username in received message: got %v, want %v", receivedMessage.Username, mockMessage.Username)
		}
	}))
	defer server.Close()

	// Connect to the mock WebSocket server
	u := "ws" + strings.TrimPrefix(server.URL, "http") + "/"
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// Wait for a short time to allow the HandleMessages goroutine to process the message
	time.Sleep(time.Millisecond * 100)
}

func TestServeChat_Unauthenticated(t *testing.T) {
	mockToken := "invalid-token"
	mockTemplate := template.New("")
	handler := ServeChat(mockTemplate)

	req, err := http.NewRequest("GET", "/chat?token="+mockToken, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Validate the redirect location
	expectedRedirectURL := "/login"
	if location := rr.Header().Get("Location"); location != expectedRedirectURL {
		t.Errorf("handler returned wrong redirect location: got %v want %v", location, expectedRedirectURL)
	}
}

func TestCallBotAPI(t *testing.T) {
	// Create a mock server to simulate the external API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Mock API response")
	}))
	defer mockServer.Close()

	// Replace the URL with the mock server URL
	mockURL := mockServer.URL + "/stock-quote?stock_code=TEST"
	callBotAPI(mockURL)
}
