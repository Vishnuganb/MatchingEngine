package handler

import (
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
	SaveOrderAndEvent(ctx context.Context, order orderBook.Order, event orderBook.Event) (orderBook.Order, orderBook.Event, error)
	UpdateOrderAndEvent(ctx context.Context, orderID string, leavesQty decimal.Decimal, event orderBook.Event) error
	CancelEvent(ctx context.Context, event orderBook.Event) error
}

type OrderBook interface {
	NewOrder(order orderBook.Order) orderBook.Event
	CancelOrder(orderID string) orderBook.Event
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
		h.handleNewOrder(ctx, msg, req)

	case rmq.ReqTypeCancel:
		h.handleCancelOrder(ctx, msg, req)

	default:
		h.handleFailure(msg, fmt.Errorf("unknown request type: %v", req.RequestType))
		return
	}
	log.Printf("Processed request of type %d in %v", req.RequestType, time.Since(start))
}

func (h *OrderRequestHandler) handleNewOrder(ctx context.Context, msg amqp.Delivery, req rmq.OrderRequest) {
	order := orderBook.Order{
		ID:         req.Order.ID,
		Price:      decimal.RequireFromString(req.Order.Price),
		Qty:        decimal.RequireFromString(req.Order.Qty),
		Instrument: req.Order.Instrument,
		Timestamp:  time.Now().UnixNano(),
		IsBid:      req.Order.Side == orderBook.Buy,
	}

	initialEvent := orderBook.Event{
		ID:         order.ID,
		OrderID:    order.ID,
		Timestamp:  time.Now().UnixNano(),
		Type:       orderBook.EventTypeNew,
		Side:       order.Side(),
		OrderQty:   order.Qty,
		LeavesQty:  order.Qty,
		Price:      order.Price,
		Instrument: order.Instrument,
	}

	savedOrder, savedEvent, err := h.OrderService.SaveOrderAndEvent(ctx, order, initialEvent)
	if err != nil {
		h.handleServiceError(msg, err, "failed to save order and event")
		return
	}

	event := h.OrderBook.NewOrder(savedOrder)

	updatedEvent := orderBook.Event{
		ID:         savedEvent.ID,
		OrderID:    savedOrder.ID,
		Timestamp:  time.Now().UnixNano(),
		Type:       event.Type,
		OrderQty:   event.OrderQty,
		LeavesQty:  event.LeavesQty,
		ExecQty:    event.ExecQty,
		Price:      event.Price,
		Instrument: event.Instrument,
	}

	err = h.OrderService.UpdateOrderAndEvent(ctx, savedOrder.ID, savedOrder.LeavesQty, updatedEvent)
	if err != nil {
		h.handleServiceError(msg, err, "failed to update order and event")
		return
	}

	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
}

func (h *OrderRequestHandler) handleCancelOrder(ctx context.Context, msg amqp.Delivery, req rmq.OrderRequest) {
	canceledEvent := h.OrderBook.CancelOrder(req.Order.ID)

	log.Println("CanceledEvent", canceledEvent)

	err := h.OrderService.CancelEvent(ctx, canceledEvent)
	if err != nil {
		h.handleServiceError(msg, err, "failed to update canceled order and event")
		return
	}

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
