package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/shopspring/decimal"
	"github.com/streadway/amqp"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

type OrderService interface {
	SaveOrderAsync(order model.Order)
	UpdateOrderAsync(orderID, orderStatus string, leavesQty, cumQty, price decimal.Decimal)
}

type OrderBook interface {
	OnNewOrder(order orderBook.Order)
	CancelOrder(orderID string)
}

type TradeNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderRequestHandler struct {
	orderBooks    map[string]OrderBook
	orderChannels map[string]chan rmq.OrderRequest
	OrderService  OrderService
	TradeNotifier TradeNotifier
	mu            sync.Mutex
}

func NewOrderRequestHandler(orderService OrderService, tradeNotifier TradeNotifier) *OrderRequestHandler {
	return &OrderRequestHandler{
		orderBooks:    make(map[string]OrderBook),
		orderChannels: make(map[string]chan rmq.OrderRequest),
		OrderService:  orderService,
		TradeNotifier: tradeNotifier,
	}
}

func (h *OrderRequestHandler) HandleOrderMessage(msg amqp.Delivery) {
	var req rmq.OrderRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		h.handleFailure(msg, fmt.Errorf("invalid message format: %w", err))
		return
	}
	// Get or create the order channel for the instrument
	h.mu.Lock()
	orderChannel, exists := h.orderChannels[req.Order.Instrument]
	if !exists {
		orderChannel = make(chan rmq.OrderRequest, 100)
		h.orderChannels[req.Order.Instrument] = orderChannel
		go h.startOrderWorkerForInstrument(req.Order.Instrument, orderChannel)
	}
	h.mu.Unlock()

	// Send the order to the instrument-specific channel
	select {
	case orderChannel <- req:
		// Successfully Added to the already running worker
	default:
		log.Printf("Order channel for instrument %s is full, dropping order: %v", req.Order.Instrument, req)
	}

	// Acknowledge a message after successful processing
	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
}

func (h *OrderRequestHandler) startOrderWorkerForInstrument(instrument string, orderChannel chan rmq.OrderRequest) {
	for req := range orderChannel {
		// Get or create the order book for the instrument
		h.mu.Lock()
		book, exists := h.orderBooks[instrument]
		if !exists {
			book = orderBook.NewOrderBook(h.TradeNotifier)
			h.orderBooks[instrument] = book
		}
		h.mu.Unlock()

		// Process the request
		switch req.RequestType {
		case rmq.ReqTypeNew:
			h.handleNewOrder(book, req)
		case rmq.ReqTypeCancel:
			h.handleCancelOrder(book, req)
		default:
			log.Printf("Unknown request type: %v", req.RequestType)
		}
	}
}

func (h *OrderRequestHandler) handleNewOrder(book OrderBook, req rmq.OrderRequest) {
	internalOrder := toInternalOrder(req)
	book.OnNewOrder(internalOrder)
}

func (h *OrderRequestHandler) handleCancelOrder(book OrderBook, req rmq.OrderRequest) {
	book.CancelOrder(req.Order.ID)
}

func (h *OrderRequestHandler) HandleExecutionReport(message []byte) error {
	var event model.ExecutionReport
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Error unmarshaling JSON: %v, message: %s", err, string(message))
		return nil // Skip processing this message
	}

	switch event.ExecType {
	case string(orderBook.ExecTypeNew), string(orderBook.ExecTypePendingNew):
		h.OrderService.SaveOrderAsync(h.convertEventToOrder(event))
	case string(orderBook.ExecTypeFill), string(orderBook.ExecTypeCanceled), string(orderBook.ExecTypeRejected):
		h.OrderService.UpdateOrderAsync(
			event.OrderID,
			event.OrderStatus,
			event.LeavesQty,
			event.CumQty,
			event.Price,
		)
	default:
		return fmt.Errorf("unknown execution type: %s", event.ExecType)
	}

	return nil
}

func (h *OrderRequestHandler) handleFailure(msg amqp.Delivery, err error) {
	log.Printf("Message failed: %v, error: %v", string(msg.Body), err)
	if err := msg.Nack(false, false); err != nil {
		log.Printf("Failed to negatively acknowledge message: %v", err)
	}
}

func (s *OrderRequestHandler) convertEventToOrder(execution model.ExecutionReport) model.Order {
	return model.Order{
		ID:          execution.OrderID,
		Instrument:  execution.Instrument,
		Price:       execution.Price,
		OrderQty:    execution.OrderQty,
		LeavesQty:   execution.LeavesQty,
		CumQty:      execution.CumQty,
		IsBid:       execution.IsBid,
		OrderStatus: execution.OrderStatus,
	}
}
