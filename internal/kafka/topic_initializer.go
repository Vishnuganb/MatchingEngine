package kafka

import (
	"log"

	"github.com/segmentio/kafka-go"
)

func InitializeTopics(brokerAddr string, topics []string) error {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, topic := range topics {
		// Check if the topic already exists
		partitions, err := conn.ReadPartitions()
		if err != nil {
			log.Printf("Failed to read partitions: %v", err)
			return err
		}

		topicExists := false
		for _, p := range partitions {
			if p.Topic == topic {
				topicExists = true
				break
			}
		}

		if topicExists {
			log.Printf("Topic %s already exists, skipping creation", topic)
			continue
		}

		// Create the topic if it doesn't exist
		err = conn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
		if err != nil {
			log.Printf("Failed to create topic %s: %v", topic, err)
			return err
		}
		log.Printf("Topic %s created successfully", topic)
	}
	return nil
}