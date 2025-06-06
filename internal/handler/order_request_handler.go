package handler

import (
	"MatchingEngine/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"github.com/streadway/amqp"

	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderService interface {
	SaveOrderAsync(order model.Order)
	SaveEventAsync(event model.Event)
	CancelEventAsync(event model.Event)
}

type OrderBook interface {
	OnNewOrder(order model.Order) model.Events
	CancelOrder(orderID string) model.Event
}

type OrderRequestHandler struct {
	OrderBook    OrderBook
	OrderService OrderService
}

func NewOrderRequestHandler(orderBook OrderBook, orderService OrderService) *OrderRequestHandler {
	return &OrderRequestHandler{
		OrderBook:    orderBook,
		OrderService: orderService,
	}
}

func (h *OrderRequestHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) {
	start := time.Now()
	var req rmq.OrderRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		h.handleFailure(msg, fmt.Errorf("invalid message format: %w", err))
		return
	}
	switch req.RequestType {
	case rmq.ReqTypeNew:
		h.handleNewOrder(msg, req)

	case rmq.ReqTypeCancel:
		h.handleCancelOrder(msg, req)

	default:
		h.handleFailure(msg, fmt.Errorf("unknown request type: %v", req.RequestType))
		return
	}
	log.Printf("Processed request of type %d in %v", req.RequestType, time.Since(start))
}

func (h *OrderRequestHandler) handleNewOrder(msg amqp.Delivery, req rmq.OrderRequest) {
	order := model.Order{
		ID:         req.Order.ID,
		Price:      decimal.RequireFromString(req.Order.Price),
		Qty:        decimal.RequireFromString(req.Order.Qty),
		Instrument: req.Order.Instrument,
		Timestamp:  time.Now().UnixNano(),
		IsBid:      req.Order.Side == orderBook.Buy,
	}

	// Save the order and generate events
	h.OrderService.SaveOrderAsync(order)

	events := h.OrderBook.OnNewOrder(order)

	// Save each event
	for _, event := range events {
		h.OrderService.SaveEventAsync(event)
	}

	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
}

func (h *OrderRequestHandler) handleCancelOrder(msg amqp.Delivery, req rmq.OrderRequest) {
	canceledEvent := h.OrderBook.CancelOrder(req.Order.ID)

	log.Println("CanceledEvent", canceledEvent)

	h.OrderService.CancelEventAsync(canceledEvent)

	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
}

func (h *OrderRequestHandler) handleServiceError(msg amqp.Delivery, err error, contextMsg string) {
	if errors.Is(err, ErrOrderNotFound) {
		log.Printf("Business error: %v", err)
		if err := msg.Ack(false); err != nil {
			log.Printf("Failed to acknowledge message: %v", err)
		}
		return
	}
	h.handleFailure(msg, fmt.Errorf("%s: %w", contextMsg, err))
}

func (h *OrderRequestHandler) handleFailure(msg amqp.Delivery, err error) {
	log.Printf("Message failed: %v, error: %v", string(msg.Body), err)
	if err := msg.Nack(false, false); err != nil {
		log.Printf("Failed to negatively acknowledge message: %v", err)
	}
}
