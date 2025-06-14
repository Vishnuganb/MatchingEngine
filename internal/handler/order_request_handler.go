package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/streadway/amqp"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

type OrderService interface {
	SaveOrderAsync(order model.Order)
	UpdateOrderAsync(orderID, orderStatus, execType string, leavesQty, execQty decimal.Decimal)
}

type OrderBook interface {
	OnNewOrder(order model.Order, producer EventNotifier)
	CancelOrder(orderID string, producer EventNotifier)
}

type EventNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderRequestHandler struct {
	orderBooks    map[string]*orderBook.OrderBook
	orderChannels map[string]chan rmq.OrderRequest
	OrderService  OrderService
	eventNotifier EventNotifier // Add this field
	mu            sync.Mutex
	orderLocks    sync.Map
}

func NewOrderRequestHandler(orderService OrderService, eventNotifier EventNotifier) *OrderRequestHandler {
	return &OrderRequestHandler{
		orderBooks:    make(map[string]*orderBook.OrderBook),
		orderChannels: make(map[string]chan rmq.OrderRequest),
		OrderService:  orderService,
		eventNotifier: eventNotifier,
	}
}

func (h *OrderRequestHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) {
	var req rmq.OrderRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		h.handleFailure(msg, fmt.Errorf("invalid message format: %w", err))
		return
	}
	// Get or create a lock for the order ID
	// It returns the value associated with the key
	lock, _ := h.orderLocks.LoadOrStore(req.Order.ID, &sync.Mutex{})
	orderLock := lock.(*sync.Mutex)
	orderLock.Lock()
	defer func() {
		orderLock.Unlock()
		h.orderLocks.Delete(req.Order.ID)
	}()

	// Get or create the order book for the instrument
	h.mu.Lock()
	book, exists := h.orderBooks[req.Order.Instrument]
	if !exists {
		book = orderBook.NewOrderBook(h.eventNotifier)
		h.orderBooks[req.Order.Instrument] = book
	}
	h.mu.Unlock()

	// Process the request immediately (synchronously)
	switch req.RequestType {
	case rmq.ReqTypeNew:
		h.handleNewOrder(book, req)
	case rmq.ReqTypeCancel:
		h.handleCancelOrder(book, req)
	default:
		log.Printf("Unknown request type: %v", req.RequestType)
	}

	// Acknowledge a message after successful processing
	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
}

func (h *OrderRequestHandler) handleNewOrder(book *orderBook.OrderBook, req rmq.OrderRequest) {
	order := model.Order{
		ID:          req.Order.ID,
		Price:       decimal.RequireFromString(req.Order.Price),
		OrderQty:    decimal.RequireFromString(req.Order.Qty),
		Instrument:  req.Order.Instrument,
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: string(orderBook.EventTypePendingNew),
		IsBid:       req.Order.Side == orderBook.Buy,
	}

	book.OnNewOrder(order, book.KafkaProducer)
}

func (h *OrderRequestHandler) handleCancelOrder(book *orderBook.OrderBook, req rmq.OrderRequest) {
	book.CancelOrder(req.Order.ID, book.KafkaProducer)
}

func (h *OrderRequestHandler) HandleEventMessages(message []byte) error {
	var event model.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Error unmarshaling JSON: %v, message: %s", err, string(message))
		return nil // Skip processing this message
	}

	switch event.EventType {
	case string(orderBook.EventTypeNew), string(orderBook.EventTypePendingNew):
		h.OrderService.SaveOrderAsync(h.convertEventToOrder(event))
	case string(orderBook.EventTypeFill), string(orderBook.EventTypePartialFill),
		string(orderBook.EventTypeCanceled), string(orderBook.EventTypeRejected):
		h.OrderService.UpdateOrderAsync(
			event.OrderID,
			event.OrderStatus,
			event.ExecType,
			event.LeavesQty,
			event.ExecQty,
		)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}

	return nil
}

func (h *OrderRequestHandler) handleFailure(msg amqp.Delivery, err error) {
	log.Printf("Message failed: %v, error: %v", string(msg.Body), err)
	if err := msg.Nack(false, false); err != nil {
		log.Printf("Failed to negatively acknowledge message: %v", err)
	}
}

func (s *OrderRequestHandler) convertEventToOrder(event model.OrderEvent) model.Order {
	return model.Order{
		ID:          event.OrderID,
		Instrument:  event.Instrument,
		Price:       event.Price,
		OrderQty:    event.Quantity,
		LeavesQty:   event.LeavesQty,
		ExecQty:     event.ExecQty,
		IsBid:       event.IsBid,
		OrderStatus: event.OrderStatus,
		ExecType:    event.ExecType,
	}
}
