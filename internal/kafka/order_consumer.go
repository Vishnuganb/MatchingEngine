package kafka

import (
	"context"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type ConsumerOpts struct {
	BrokerAddrs string
	Topic       string
	GroupID     string
}

type Consumer struct {
	opts           ConsumerOpts
	reader         *kafka.Reader
	requestHandler MessageHandler
	batch          *MessageBatch
}

type MessageHandler interface {
	HandleExecutionReport(message []byte) error
}

func NewConsumer(opts ConsumerOpts, requestHandler MessageHandler) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{opts.BrokerAddrs},
		Topic:       opts.Topic,
		GroupID:     opts.GroupID,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		opts:           opts,
		reader:         reader,
		requestHandler: requestHandler,
		batch:          NewMessageBatch(),
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Kafka consumer started for topic:%s, groupID:%s", c.opts.Topic, c.opts.GroupID)

	// Start a message collection goroutine
	go c.collectMessages(ctx)

	// Start periodic processing of collected messages
	go c.processBatchPeriodically(ctx)

	<-ctx.Done()
	return nil
}

func (c *Consumer) collectMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping message collection due to context cancellation")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("Context cancelled, stopping message collection")
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			// Add a message to the batch instead of processing immediately
			c.batch.AddMessage(msg.Value)

			// commit the message to mark it as processed
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

func (c *Consumer) processBatchPeriodically(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping batch processing due to context cancellation")
			return
		case <-ticker.C:
			messages := c.batch.GetAndClearMessages()
			if len(messages) > 0 {
				log.Printf("Processing batch of %d messages", len(messages))
				for _, msg := range messages {
					if err := c.requestHandler.HandleExecutionReport(msg); err != nil {
						log.Printf("Error handling message: %v", err)
					}
				}
			}
		}
	}
}
