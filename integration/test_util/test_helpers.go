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

func ConsumeKafkaMessages(topic string) (<-chan string, func()) {
	brokers := os.Getenv("KAFKA_BROKER")
	groupID := fmt.Sprintf("test-consumer-%s-%d", topic, time.Now().UnixNano())

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{brokers},
		Topic:          topic,
		GroupID:        groupID,
		CommitInterval: time.Second,
		ReadBackoffMin: time.Millisecond * 100,
		ReadBackoffMax: time.Second * 1,
	})

	messageChan := make(chan string)
	done := make(chan struct{})

	cleanup := func() {
		close(done)
		reader.Close()
	}

	go func() {
		defer close(messageChan)
		for {
			select {
			case <-done:
				return
			default:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				msg, err := reader.ReadMessage(ctx)
				cancel()

				if err != nil {
					if err != context.DeadlineExceeded {
						log.Printf("Error reading Kafka message: %v", err)
					}
					continue
				}

				select {
				case messageChan <- string(msg.Value):
				case <-done:
					return
				}
			}
		}
	}()

	return messageChan, cleanup
}
