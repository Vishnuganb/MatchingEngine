package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/repository"
	"MatchingEngine/orderBook"
)

type OrderService struct {
	repo          repository.OrderRepository
	KafkaProducer Producer
}

type Producer interface {
	NotifyEventAndOrder(key string, value json.RawMessage) error
}

func NewOrderService(repo repository.OrderRepository, kafkaProducer Producer) *OrderService {
	return &OrderService{
		repo:          repo,
		KafkaProducer: kafkaProducer,
	}
}

func (s *OrderService) SaveOrderAndEvent(ctx context.Context, order orderBook.Order, event orderBook.Event) (orderBook.Order, orderBook.Event, error) {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	order, err := s.repo.SaveOrder(ctx, order)
	if err != nil {
		log.Println("Failed to create order:", err)
		return orderBook.Order{}, orderBook.Event{}, err
	}

	event, err = s.repo.SaveEvent(ctx, event)
	if err != nil {
		log.Println("Failed to create event:", err)
		return orderBook.Order{}, orderBook.Event{}, err
	}

	s.pushEvents(event)

	return order, event, nil
}

func (s *OrderService) UpdateOrderAndEvent(ctx context.Context, orderID string, leavesQty decimal.Decimal, event orderBook.Event) error {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Update the order
	_, err := s.repo.UpdateOrder(ctx, orderID, leavesQty)
	if err != nil {
		log.Println("Failed to update order:", err)
		return err
	}

	// Update the event
	updatedEvent, err := s.repo.UpdateEvent(ctx, event)
	if err != nil {
		log.Println("Failed to update event:", err)
		return err
	}

	s.pushEvents(updatedEvent)

	return nil
}

func (s *OrderService) CancelEvent(ctx context.Context, event orderBook.Event) error {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Update the event for cancellation
	canceledEvent, err := s.repo.SaveEvent(ctx, event)
	if err != nil {
		log.Println("Failed to update canceled event:", err)
		return err
	}

	s.pushEvents(canceledEvent)

	return nil
}

func (s *OrderService) pushEvents(event orderBook.Event) {
	// Serialize the event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to serialize event: %v", err)
		return
	}

	log.Printf("Pushing event: %v", string(eventJSON))

	// Publish the event to Kafka
	err = s.KafkaProducer.NotifyEventAndOrder(event.ID, eventJSON)
	if err != nil {
		log.Printf("Failed to publish event to Kafka: %v", err)
		return
	}
}
