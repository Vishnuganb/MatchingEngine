package service

import (
	"context"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/repository"
	"MatchingEngine/orderBook"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
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

	return order, event, nil
}

func (s *OrderService) UpdateOrderAndEvent(ctx context.Context, orderID string, leavesQty decimal.Decimal, event orderBook.Event) (orderBook.Order, orderBook.Event, error) {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Update the order
	order, err := s.repo.UpdateOrder(ctx, orderID, leavesQty)
	if err != nil {
		log.Println("Failed to update order:", err)
		return orderBook.Order{}, orderBook.Event{}, err
	}

	// Update the event
	updatedEvent, err := s.repo.UpdateEvent(ctx, event)
	if err != nil {
		log.Println("Failed to update event:", err)
		return orderBook.Order{}, orderBook.Event{}, err
	}

	return order, updatedEvent, nil
}
