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
		Brokers: []string{brokers},
		Topic:   topic,
		GroupID: groupID,
	})
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		_, err := reader.ReadMessage(ctx)
		if err != nil {
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

	messageChan := make(chan string, 100)

	go func() {
		defer close(messageChan)
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
