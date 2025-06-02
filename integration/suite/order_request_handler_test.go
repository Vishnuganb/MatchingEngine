package suite

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/streadway/amqp"

	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

func TestOrderRequestHandler_Integration(t *testing.T) {
	// Setup RabbitMQ connection
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	queueName := "order_requests"
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		t.Fatalf("Failed to declare a queue: %v", err)
	}

	// Create a mock OrderRequest
	orderRequest := rmq.OrderRequest{
		RequestType: rmq.ReqTypeNew,
		Order: rmq.TraderOrder{
			ID:    "order-id-3",
			Side:  orderBook.Buy,
			Instrument: "BTC/USDT",
			Qty:   "105.0",
			Price: "508.0",
		},
	}
	body, err := json.Marshal(orderRequest)
	if err != nil {
		t.Fatalf("Failed to marshal order request: %v", err)
	}

	// Publish the message to the queue
	err = ch.Publish("", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		t.Fatalf("Failed to publish a message: %v", err)
	}

	// Wait for the message to be processed
	time.Sleep(2 * time.Second)
}
