package test_util

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
)

func SetupRabbitMQConnection() *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	return conn
}

func PublishOrder(ch *amqp.Channel, queueName string, order []byte) {
	err := ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        order,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish order: %v", err)
	}
}

func ClearKafkaTopic(topic string) {
	brokers := os.Getenv("KAFKA_BROKER")

	// Create a reader with a unique group ID to ensure we get all messages
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokers},
		Topic:       topic,
		GroupID:     fmt.Sprintf("cleanup-%d", time.Now().UnixNano()),
		StartOffset: kafka.FirstOffset,
		// Set a larger MinBytes and MaxBytes to fetch more messages at once
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader during cleanup: %v", err)
		}
	}()

	// Use a longer timeout for cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messagesCleared := 0
	for {
		// Read messages in batches until we hit an error or timeout
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded || err == io.EOF {
				log.Printf("Finished clearing topic %s: cleared %d messages", topic, messagesCleared)
				break
			}
			log.Printf("Error reading message during cleanup: %v", err)
			break
		}
		if msg.Value != nil {
			messagesCleared++
		}

		// Commit the offset after reading each message
		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Error committing message offset: %v", err)
		}
	}

	// Double-check if there are any remaining messages
	for {
		_, err := reader.ReadMessage(ctx)
		if err != nil {
			break
		}
		messagesCleared++
	}

	log.Printf("Topic cleanup completed. Cleared %d messages total", messagesCleared)
}

func ConsumeKafkaMessages(topic string) <-chan string {
	brokers := os.Getenv("KAFKA_BROKER")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   topic,
		GroupID: "suite.test", // Use consistent group ID
		// Start from beginning of topic
		StartOffset: kafka.FirstOffset,
	})

	messageChan := make(chan string, 100)

	// Increase timeout to 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	go func() {
		defer func() {
			cancel()
			if err := reader.Close(); err != nil {
				log.Printf("Error closing reader: %v", err)
			}
			close(messageChan)
		}()

		for {
			select {
			case <-ctx.Done():
				log.Printf("Consumer context done: %v", ctx.Err())
				return
			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					log.Printf("Error reading message: %v", err)
					if err == context.DeadlineExceeded {
						return
					}
					time.Sleep(100 * time.Millisecond) // Add small delay on error
					continue
				}
				log.Printf("Consumer received message: %s", string(msg.Value))
				messageChan <- string(msg.Value)
			}
		}
	}()

	return messageChan
}
