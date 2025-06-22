package handler

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"

	"MatchingEngine/internal/model"
)

type ExecutionService interface {
	SaveExecutionAsync(order model.ExecutionReport)
}

type TradeService interface {
	SaveTradeAsync(trade model.Trade)
}

type OrderService interface {
	ProcessOrderRequest(req model.OrderRequest) error
}

type OrderRequestHandler struct {
	OrderService     OrderService
	ExecutionService ExecutionService
	TradeService     TradeService
}

func NewOrderRequestHandler(executionService ExecutionService, tradeService TradeService, orderService OrderService) *OrderRequestHandler {
	return &OrderRequestHandler{
		OrderService:     orderService,
		ExecutionService: executionService,
		TradeService:     tradeService,
	}
}

func (h *OrderRequestHandler) HandleOrderMessage(msg amqp.Delivery) {
	var req model.OrderRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		h.handleFailure(msg, fmt.Errorf("invalid message format: %w", err))
		return
	}

	err := h.OrderService.ProcessOrderRequest(req)
	if err != nil {
		log.Printf("Failed to process order request: %v, message: %s", err, string(msg.Body))
		h.handleFailure(msg, err)
		return
	}

	// Acknowledge a message after successful processing
	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
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
