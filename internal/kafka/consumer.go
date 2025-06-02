package kafka

import (
	"context"
	"log"

	kafka "github.com/segmentio/kafka-go"
)

func StartConsumer(brokerAddr, topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic:   topic,
		GroupID: "order-group",
	})

	go func() {
		defer func() {
			if err := reader.Close(); err != nil {
				log.Printf("Error closing Kafka reader: %v", err)
			}
		}()
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Println("Kafka consumer error:", err)
				continue
			}

			// Log the consumed message
			log.Printf("Consumed message: key=%s, value=%s", string(m.Key), string(m.Value))

		}
	}()
}
