package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafka "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

type EventNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

func NewProducer(brokerAddr string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddr),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) NotifyEventAndTrade(key string, value json.RawMessage) error {
	// Serialize the value to JSON
	log.Printf("Publishing message to Kafka: Key=%s, Value=%s", key, string(value))
	err := p.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: value,
	})
	if err != nil {
		log.Println("failed to publish message:", err)
	}
	return err
}
