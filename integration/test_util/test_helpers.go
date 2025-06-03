package test_util

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
	"log"
	"time"
)

func SetupRabbitMQConnection() *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

// SetupKafkaConsumer sets up a Kafka consumer for testing purposes
func SetupKafkaConsumer(brokerAddr, topic, groupID string) *kafka.Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic:   topic,
		GroupID: groupID,
	})

	return reader
}

// ConsumeMessages consumes messages from the Kafka topic
func ConsumeMessages(reader *kafka.Reader, timeout time.Duration) []string {
	var messages []string
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				break
			}
			log.Printf("Error reading message: %v", err)
			continue
		}
		messages = append(messages, string(m.Value))
	}

	return messages
}
