package handler

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"

	"MatchingEngine/internal/model"
)

type ExecutionService interface {
	SaveExecutionAsync(order model.ExecutionReport)
}

type TradeService interface {
	SaveTradeAsync(trade model.TradeCaptureReport)
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
		log.Printf("failed to decode order request: %v | message: %s", err, string(msg.Body))
		h.handleFailure(msg, "invalid JSON format")
		return
	}

	log.Printf("Received order request: %+v", req)

	err := h.OrderService.ProcessOrderRequest(req)
	if err != nil {
		log.Printf("failed to process order request: %v | message: %s", err, string(msg.Body))
		h.handleFailure(msg, "order processing error")
		return
	}

	// Acknowledge a message after successful processing
	if err := msg.Ack(false); err != nil {
		log.Printf("failed to acknowledge message: %v", err)
	}
}

func (h *OrderRequestHandler) HandleExecutionReport(message []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		log.Printf("invalid execution report JSON: %v | message: %s", err, string(message))
		return nil // Skip processing this message
	}

	msgType, ok := raw["35"].(string)
	if !ok {
		log.Printf("missing or invalid MsgType in message: %s", string(message))
		return nil
	}

	switch msgType {
	case string(model.MsgTypeTradeReport):
		var tradeCaptureReport model.TradeCaptureReport
		if err := h.unmarshalAndLogError(message, &tradeCaptureReport); err != nil {
			return nil
		}
		log.Printf("received trade report : %+v", tradeCaptureReport)
		h.TradeService.SaveTradeAsync(tradeCaptureReport)

	case string(model.MsgTypeExecRpt):
		var execReport model.ExecutionReport
		if err := h.unmarshalAndLogError(message, &execReport); err != nil {
			return nil
		}
		log.Printf("received execution report: %+v", execReport)
		h.ExecutionService.SaveExecutionAsync(execReport)

	default:
		log.Printf("unknown MsgType: %s | message: %s", msgType, string(message))
	}

	return nil
}

func (h *OrderRequestHandler) unmarshalAndLogError(message []byte, v interface{}) error {
	if err := json.Unmarshal(message, v); err != nil {
		log.Printf("failed to parse FIX message: %v | message: %s", err, string(message))
		return err
	}
	return nil
}

func (h *OrderRequestHandler) handleFailure(msg amqp.Delivery, reason string) {
	log.Printf("nacking message due to: %s | body: %s", reason, string(msg.Body))
	if err := msg.Nack(false, false); err != nil {
		log.Printf("failed to negatively acknowledge message: %v", err)
	}
}
