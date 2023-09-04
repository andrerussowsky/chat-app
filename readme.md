# Chat App

This is a simple browser-based chat application implemented in Go. It allows users to communicate with each other in real-time using websockets.

## Getting Started

### Prerequisites

- Go (version 1.13 or higher) (https://golang.org/doc/install)
- Docker (https://docs.docker.com/get-docker/)
- Internet connection (for stock quote retrieval)

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/andrerussowsky/chat-app.git

2. Change into the project directory:

   ```sh
   cd chat-app

3. Start RabbitMQ using Docker Compose:

   ```sh
   docker-compose up -d

This will start a RabbitMQ container and expose ports 5672 and 15672 (for the RabbitMQ Management UI).

3. Install project dependencies:

   ```sh
   go mod tidy

### Usage

1. Run the chat application server:

   ```sh
   go run cmd/main.go

The server should start and listen on port 8080.

2. Open your web browser and go to http://localhost:8080 to access the chat application.

3. Register a new account or log in with an existing one.

4. Start chatting with other users in real-time!

### Running the Bot

1. Open a new terminal and change into the bot directory:

   ```sh
   cd chat-app/bot

2. Run the bot application:

   ```sh
   go run main.go

The bot will start and listen on port 8082.

3. The bot will respond to stock code commands in the chatroom (e.g., /stock=AAPL.US). It will fetch stock data using an external API, process the CSV response, and send stock quote messages back to the chatroom.

