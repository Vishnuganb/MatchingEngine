package test_util

import (
	"context"
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

// SetupKafkaConsumer sets up a Kafka consumer for testing purposes
func StartConsumer(topic string) string {
	brokers := os.Getenv("KAFKA_BROKER")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{brokers},
		Topic:         topic,
		GroupID:       "eventConsumer",
		ReadBackoffMin: time.Millisecond * 100,
		ReadBackoffMax: time.Second * 1,
	})

	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Increased timeout to 30 seconds
	defer cancel()

	m, err := reader.ReadMessage(ctx)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Printf("Timeout reached while consuming Kafka message for topic: %s", topic)
			return ""
		}
		log.Printf("Kafka consumer error for topic %s: %v", topic, err)
		return ""
	}

	log.Printf("Consumed message from topic %s: key=%s, value=%s", topic, string(m.Key), string(m.Value))
	return string(m.Value)
}

func ClearKafkaTopic(topic string) {
	brokers := os.Getenv("KAFKA_BROKER")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   topic,
		GroupID: "test-clear-group",
	})
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Set a 10-second timeout
	defer cancel()

	for {
		_, err := reader.ReadMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded {
				log.Println("Timeout reached while clearing Kafka topic")
				break
			}
			log.Printf("Error reading message: %v", err)
			break
		}
	}
}

func ConsumeKafkaMessages(topic string) <-chan string {
	brokers := os.Getenv("KAFKA_BROKER")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   topic,
		GroupID: "test-consumer-group",
	})

	messageChan := make(chan string)

	go func() {
		defer reader.Close()
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading Kafka message: %v", err)
				close(messageChan)
				return
			}
			messageChan <- string(msg.Value)
		}
	}()

	return messageChan
}
