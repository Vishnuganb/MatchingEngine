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

	"MatchingEngine/internal/kafka"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/internal/service"
	"MatchingEngine/orderBook"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRequestHandler struct {
	OrderBook     *orderBook.OrderBook
	OrderService  *service.OrderService
	KafkaProducer *kafka.Producer
}

func NewOrderRequestHandler(orderBook *orderBook.OrderBook, orderService *service.OrderService, kafkaProducer *kafka.Producer) *OrderRequestHandler {
	return &OrderRequestHandler{
		OrderBook:     orderBook,
		OrderService:  orderService,
		KafkaProducer: kafkaProducer,
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

	_, event, err = h.OrderService.UpdateOrderAndEvent(ctx, savedOrder.ID, savedOrder.LeavesQty, updatedEvent)
	if err != nil {
		h.handleServiceError(msg, err, "failed to update order and event")
		return
	}

	h.pushEvents(event)
	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
}

func (h *OrderRequestHandler) handleCancelOrder(ctx context.Context, msg amqp.Delivery, req rmq.OrderRequest) {
	canceledEvent := h.OrderBook.CancelOrder(req.Order.ID)

	log.Println("CanceledEvent",canceledEvent)

	_, event, err := h.OrderService.UpdateOrderAndEvent(ctx, req.Order.ID, decimal.Zero, canceledEvent)
	if err != nil {
		h.handleServiceError(msg, err, "failed to update canceled order and event")
		return
	}

	h.pushEvents(event)
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

func (h *OrderRequestHandler) pushEvents(event orderBook.Event) {
	// Serialize the event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to serialize event: %v", err)
		return
	}

	log.Printf("Pushing event: %v", string(eventJSON))

	// Publish the event to Kafka
	err = h.KafkaProducer.NotifyEventAndOrder(event.ID, eventJSON)
	if err != nil {
		log.Printf("Failed to publish event to Kafka: %v", err)
		return
	}
}

func (h *OrderRequestHandler) handleFailure(msg amqp.Delivery, err error) {
	log.Printf("Message failed: %v, error: %v", string(msg.Body), err)
	if err := msg.Nack(false, false); err != nil {
		log.Printf("Failed to negatively acknowledge message: %v", err)
	}
}
