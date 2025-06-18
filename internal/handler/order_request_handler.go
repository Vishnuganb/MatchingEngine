package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/streadway/amqp"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

type ExecutionService interface {
	SaveExecutionAsync(order model.ExecutionReport)
}

type TradeService interface {
	SaveTradeAsync(trade model.Trade)
}

type OrderBook interface {
	OnNewOrder(order orderBook.OrderRequest)
	CancelOrder(orderID string)
}

type TradeNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderRequestHandler struct {
	orderBooks    map[string]OrderBook
	orderChannels map[string]chan rmq.OrderRequest
	ExecutionService  ExecutionService
	TradeService  TradeService
	TradeNotifier TradeNotifier
	mu            sync.Mutex
}

func NewOrderRequestHandler(executionService ExecutionService, tradeService TradeService, tradeNotifier TradeNotifier) *OrderRequestHandler {
	return &OrderRequestHandler{
		orderBooks:    make(map[string]OrderBook),
		orderChannels: make(map[string]chan rmq.OrderRequest),
		ExecutionService:  executionService,
		TradeService:  tradeService,
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
	orderChannel := h.getOrderChannel(req.Order.Instrument)

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

func (h *OrderRequestHandler) getOrderChannel(instrument string) chan rmq.OrderRequest {
	h.mu.Lock()
	defer h.mu.Unlock()

	orderChannel, exists := h.orderChannels[instrument]
	if !exists {
		orderChannel = make(chan rmq.OrderRequest, 100)
		h.orderChannels[instrument] = orderChannel
		go h.startOrderWorkerForInstrument(instrument, orderChannel)
	}

	return orderChannel
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
	internalOrder := toInternalOrderRequest(req)
	book.OnNewOrder(internalOrder)
}

func (h *OrderRequestHandler) handleCancelOrder(book OrderBook, req rmq.OrderRequest) {
	book.CancelOrder(req.Order.ID)
}

func (h *OrderRequestHandler) HandleExecutionReport(message []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		log.Printf("Error unmarshaling JSON: %v, message: %s", err, string(message))
		return nil // Skip processing this message
	}

	// Handle Trade
	if _, ok := raw["buyer_order_id"]; ok {
		var trade model.Trade
		if err := h.unmarshalAndLogError(message, &trade); err != nil {
			return err
		}
		log.Printf("Received trade: %+v", trade)
		h.TradeService.SaveTradeAsync(trade)
		return nil
	}

	// Handle ExecutionReport
	var execReport model.ExecutionReport
	if err := h.unmarshalAndLogError(message, &execReport); err != nil {
		return err
	}
	log.Printf("Received execution report: %+v", execReport)
	
	h.ExecutionService.SaveExecutionAsync(execReport)

	return nil
}

func (h *OrderRequestHandler) unmarshalAndLogError(message []byte, v interface{}) error {
	if err := json.Unmarshal(message, v); err != nil {
		log.Printf("Error unmarshaling message: %v, message: %s", err, string(message))
		return fmt.Errorf("invalid message format: %w", err)
	}
	return nil
}

func (h *OrderRequestHandler) handleFailure(msg amqp.Delivery, err error) {
	log.Printf("Message failed: %v, error: %v", string(msg.Body), err)
	if err := msg.Nack(false, false); err != nil {
		log.Printf("Failed to negatively acknowledge message: %v", err)
	}
}
