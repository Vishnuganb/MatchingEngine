package test_util

import (
	"context"
	"errors"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		GroupID: fmt.Sprintf("test-group-%d", time.Now().UnixNano()),
	})

	messageChan := make(chan string, 100)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	go func() {
		defer func() {
			cancel()
			close(messageChan)
			if err := reader.Close(); err != nil {
				log.Printf("Error closing reader: %v", err)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					if err == context.DeadlineExceeded {
						return
					}
					if !errors.Is(err, context.Canceled) {
						log.Printf("Error reading message: %v", err)
					}
					return
				}

				if len(msg.Value) == 0 {
					continue
				}

				log.Printf("Consumer received message: %s", string(msg.Value))
				select {
				case messageChan <- string(msg.Value):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return messageChan
}
