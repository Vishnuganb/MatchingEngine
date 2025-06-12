package test_util

import (
	"context"
	"fmt"
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
	groupID := fmt.Sprintf("clear-group-%s-%d", topic, time.Now().UnixNano())

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokers},
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader during cleanup: %v", err)
		}
	}()

	messagesCleared := 0
	for {
		_, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Finished clearing topic %s: %v (cleared %d messages)",
				topic, err, messagesCleared)
			break
		}
		messagesCleared++
	}
}

func ConsumeKafkaMessages(topic string) <-chan string {
	brokers := os.Getenv("KAFKA_BROKER")
	groupID := fmt.Sprintf("test-consumer-group-%d", time.Now().UnixNano())

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   topic,
		GroupID: groupID,
	})

	messageChan := make(chan string, 100)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

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
				return
			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					log.Printf("Error reading message: %v", err)
					return
				}
				messageChan <- string(msg.Value)
			}
		}
	}()

	return messageChan
}
