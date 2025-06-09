package handler

import (
	"MatchingEngine/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
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
	orderBooks    map[string]*orderBook.OrderBook
	orderChannels map[string]chan rmq.OrderRequest
	OrderService OrderService
	mu           sync.Mutex
}

func NewOrderRequestHandler(orderService OrderService) *OrderRequestHandler {
	return &OrderRequestHandler{
		orderBooks:    make(map[string]*orderBook.OrderBook),
		orderChannels: make(map[string]chan rmq.OrderRequest),
		OrderService: orderService,
	}
}

func (h *OrderRequestHandler) HandleMessage(ctx context.Context, msg amqp.Delivery) {
	var req rmq.OrderRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		h.handleFailure(msg, fmt.Errorf("invalid message format: %w", err))
		return
	}
	h.mu.Lock()
	channel, exists := h.orderChannels[req.Order.Instrument]
	_, exists = h.orderBooks[req.Order.Instrument]
	if !exists {
		channel = make(chan rmq.OrderRequest, 100) // Buffer size of 100
		h.orderChannels[req.Order.Instrument] = channel

		// Start a worker for the instrument
		go h.startWorker(req.Order.Instrument, channel)
	}
	h.mu.Unlock()

	// Send the request to the channel
	channel <- req
	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
}

func (h *OrderRequestHandler) startWorker(instrument string, channel chan rmq.OrderRequest) {
	for req := range channel {
		h.mu.Lock()
		book, exists := h.orderBooks[instrument]
		if !exists {
			book = orderBook.NewOrderBook()
			h.orderBooks[instrument] = book
		}
		h.mu.Unlock()

		switch req.RequestType {
		case rmq.ReqTypeNew:
			h.handleNewOrder(book, req)
		case rmq.ReqTypeCancel:
			h.handleCancelOrder( book, req)
		default:
			log.Printf("Unknown request type: %v", req.RequestType)
		}
		log.Printf("Processed request of type %d", req.RequestType)
	}
}

func (h *OrderRequestHandler) handleNewOrder(book *orderBook.OrderBook, req rmq.OrderRequest) {
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

	events := book.OnNewOrder(order)

	// Save each event
	for _, event := range events {
		h.OrderService.SaveEventAsync(event)
	}
}

func (h *OrderRequestHandler) handleCancelOrder(book *orderBook.OrderBook, req rmq.OrderRequest) {
	canceledEvent := book.CancelOrder(req.Order.ID)

	log.Println("CanceledEvent", canceledEvent)

	h.OrderService.CancelEventAsync(canceledEvent)
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
