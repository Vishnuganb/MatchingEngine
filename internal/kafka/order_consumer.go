package kafka

import (
	"MatchingEngine/internal/model"
	"context"
	"encoding/json"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type ConsumerOpts struct {
	BrokerAddrs string
	Topic       string
	GroupID     string
}

type MessageHandler interface {
	HandleExecutionReport(message []byte) error
}

type ExecutionService interface {
	SaveExecutionAsync(order model.ExecutionReport)
}

type TradeService interface {
	SaveTradeAsync(trade model.TradeCaptureReport)
}

type Consumer struct {
	opts         ConsumerOpts
	reader       *kafka.Reader
	executionSvc ExecutionService
	tradeSvc     TradeService
	batch        *MessageBatch
}

func NewConsumer(opts ConsumerOpts, executionService ExecutionService, tradeService TradeService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{opts.BrokerAddrs},
		Topic:       opts.Topic,
		GroupID:     opts.GroupID,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		opts:         opts,
		reader:       reader,
		executionSvc: executionService,
		tradeSvc:     tradeService,
		batch:        NewMessageBatch(),
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
					if err := c.handleReports(msg); err != nil {
						log.Printf("Error handling message: %v", err)
					}
				}
			}
		}
	}
}

func (c *Consumer) handleReports(message []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		log.Printf("Invalid execution report JSON: %v | message: %s", err, string(message))
		return nil
	}

	msgType, ok := raw["35"].(string)
	if !ok {
		log.Printf("Missing or invalid MsgType in message: %s", string(message))
		return nil
	}

	switch msgType {
	case string(model.MsgTypeTradeReport):
		var tradeCaptureReport model.TradeCaptureReport
		if err := unmarshalAndLogError(message, &tradeCaptureReport); err != nil {
			return nil
		}
		log.Printf("received trade report : %+v", tradeCaptureReport)
		c.tradeSvc.SaveTradeAsync(tradeCaptureReport)

	case string(model.MsgTypeExecRpt):
		var execReport model.ExecutionReport
		if err := unmarshalAndLogError(message, &execReport); err != nil {
			return nil
		}
		log.Printf("received execution report: %+v", execReport)
		c.executionSvc.SaveExecutionAsync(execReport)

	default:
		log.Printf("Unknown MsgType: %s | message: %s", msgType, string(message))
	}

	return nil
}

func unmarshalAndLogError(message []byte, v interface{}) error {
	if err := json.Unmarshal(message, v); err != nil {
		log.Printf("failed to parse FIX message: %v | message: %s", err, string(message))
		return err
	}
	return nil
}
