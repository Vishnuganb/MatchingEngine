package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafka "github.com/segmentio/kafka-go"
)

type Producer struct {
	dbWriter       *kafka.Writer
	executionWriter *kafka.Writer
}

type EventNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

func NewProducer(brokerAddr string, dbTopic string, executionTopic string) *Producer {
	return &Producer{
		dbWriter: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddr),
			Topic:    dbTopic,
			Balancer: &kafka.LeastBytes{},
		},
		executionWriter: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddr),
			Topic:    executionTopic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) NotifyEventAndTrade(key string, value json.RawMessage) error {
	// Publish to the database update topic
	log.Printf("Publishing message to DB topic: Key=%s, Value=%s", key, string(value))
	err := p.dbWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: value,
	})
	if err != nil {
		log.Println("Failed to publish message to DB topic:", err)
		return err
	}

	// Publish to the execution notification topic
	log.Printf("Publishing message to Execution topic: Key=%s, Value=%s", key, string(value))
	err = p.executionWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: value,
	})
	if err != nil {
		log.Println("Failed to publish message to Execution topic:", err)
		return err
	}

	return nil
}
