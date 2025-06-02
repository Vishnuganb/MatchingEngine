package suite

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"MatchingEngine/internal/handler"
	"MatchingEngine/internal/repository"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/internal/service"
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

	queueName := "test_order_requests"
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		t.Fatalf("Failed to declare a queue: %v", err)
	}

	// Create a mock OrderRequest
	orderRequest := rmq.OrderRequest{
		RequestType: rmq.ReqTypeNew,
		Order: rmq.TraderOrder{
			ID:    "test-order-id",
			Side:  orderBook.Buy,
			Qty:   "1.0",
			Price: "50000.0",
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

	// Set up dependencies for the handler
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orderBook := orderBook.NewOrderBook()
	repo := repository.NewPostgresOrderRepository(nil) // Mock or real repository
	orderService := service.NewOrderService(repo)
	kafkaProducer := &MockKafkaProducer{} // Replace with a mock Kafka producer
	handler := handler.NewOrderRequestHandler(orderBook, orderService, kafkaProducer)

	// Start the consumer
	consumerOpts := rmq.ConsumerOpts{
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
		QueueName:   queueName,
		Prefetch:    1,
	}
	consumer := rmq.NewConsumer(consumerOpts, handler.HandleMessage)

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatalf("Failed to start consumer: %v", err)
		}
	}()

	// Wait for the message to be processed
	time.Sleep(2 * time.Second)

	// Assertions (example: check if the order was processed)
	assert.NotEmpty(t, orderBook.Events, "Expected events to be generated")
	assert.Equal(t, orderBook.Events[0].OrderID, "test-order-id", "Order ID should match")
}

// MockKafkaProducer is a mock implementation of the Kafka producer
type MockKafkaProducer struct{}

func (m *MockKafkaProducer) NotifyEvent(key string, value interface{}) error {
	log.Printf("Mock Kafka event published: key=%s, value=%v", key, value)
	return nil
}
