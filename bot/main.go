package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

func main() {
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

	// Create a queue to publish messages
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

	// Listen to HTTP requests for stock code commands
	http.HandleFunc("/stock-quote", func(w http.ResponseWriter, r *http.Request) {
		stockCode := r.URL.Query().Get("stock_code")
		if stockCode != "" {
			stockQuote, err := getStockQuote(stockCode)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Publish the stock quote to RabbitMQ
			err = ch.Publish(
				"",
				q.Name,
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(stockQuote),
				},
			)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	})

	log.Fatal(http.ListenAndServe(":8082", nil))
}

func getStockQuote(stockCode string) (string, error) {
	apiURL := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", stockCode)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Println(resp.Body)

	csvData, err := ioutil.ReadAll(resp.Body)
	fmt.Println(csvData, err)
	if err != nil {
		return "", err
	}

	reader := csv.NewReader(bytes.NewReader(csvData))
	// Read the first line (header)
	_, err = reader.Read()
	if err != nil {
		return "", err
	}

	// Read the second line (data)
	record, err := reader.Read()
	if err != nil {
		return "", err
	}

	// Extract the relevant data from the CSV
	stockName := record[0]
	stockQuote := record[6] // Closing price

	return fmt.Sprintf("%s quote is $%s per share", stockName, stockQuote), nil
}
