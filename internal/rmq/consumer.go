package rmq

import (
	"context"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type ConsumerOpts struct {
	RabbitMQURL string
	QueueName   string
	Prefetch    int
}

type MessageHandler interface {
	HandleOrderMessage(msg amqp.Delivery)
}

type Consumer struct {
	opts           ConsumerOpts
	requestHandler MessageHandler
}

func NewConsumer(opts ConsumerOpts, requestHandler MessageHandler) *Consumer {
	return &Consumer{
		opts:           opts,
		requestHandler: requestHandler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	conn, err := amqp.Dial(c.opts.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		c.opts.QueueName, true, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.Qos(c.opts.Prefetch, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(
		c.opts.QueueName, "", false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				c.requestHandler.HandleOrderMessage(msg)
			}
		}
	}()

	log.Printf("Consumer started for queue: %s", c.opts.QueueName)
	<-ctx.Done()
	return nil
}
