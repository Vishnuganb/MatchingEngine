package test_util

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
	"log"
	"os"
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

// SetupKafkaConsumer sets up a Kafka consumer for testing purposes
func StartConsumer(topic string) string {
	brokers := os.Getenv("KAFKA_BROKER")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   topic,
		GroupID: "eventConsumer",
	})

	defer reader.Close()

	m, err := reader.ReadMessage(context.Background())
	if err != nil {
		log.Fatalf("Kafka consumer error: %v", err)
	}
	log.Printf("Consumed message: key=%s, value=%s", string(m.Key), string(m.Value))
	return string(m.Value)
}
