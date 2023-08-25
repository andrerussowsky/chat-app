# Chat App

This is a simple browser-based chat application implemented in Go. It allows users to communicate with each other in real-time using websockets.

## Getting Started

### Prerequisites

- Go (version 1.13 or higher)
- RabbitMQ (for bot integration)
- Internet connection (for stock quote retrieval)

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/chat-app.git

2. Change into the project directory:

cd chat-app

3. Install project dependencies:

go mod tidy

### Usage

1. Run the chat application server:

go run cmd/chatapp/main.go

The server should start and listen on port 8080.

2. Open your web browser and go to http://localhost:8080 to access the chat application.

3. Register a new account or log in with an existing one.

4. Start chatting with other users in real-time!

### Running the Bot

1. Change into the bot directory:

cd bot

2. Run the bot application:

go run main.go

The bot will start and listen on port 8082.

3. The bot will respond to stock code commands in the chatroom (e.g., /stock=AAPL.US). It will fetch stock data using an external API, process the CSV response, and send stock quote messages back to the chatroom.

