package handler

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"

	"MatchingEngine/internal/model"
)

type OrderService interface {
	ProcessOrderRequest(req model.OrderRequest) error
}

type OrderRequestHandler struct {
	OrderService     OrderService
}

func NewOrderRequestHandler(orderService OrderService) *OrderRequestHandler {
	return &OrderRequestHandler{
		OrderService:     orderService,
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

func (h *OrderRequestHandler) handleFailure(msg amqp.Delivery, reason string) {
	log.Printf("nacking message due to: %s | body: %s", reason, string(msg.Body))
	if err := msg.Nack(false, false); err != nil {
		log.Printf("failed to negatively acknowledge message: %v", err)
	}
}
