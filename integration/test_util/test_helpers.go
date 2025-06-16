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

func ConsumeKafkaMessages(topic string) <-chan string {
	brokers := os.Getenv("KAFKA_BROKER")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   topic,
		GroupID: "event_consumer_group",
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
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return // context canceled or deadline
				}
				log.Printf("Error reading message: %v", err)
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
	}()

	return messageChan
}
