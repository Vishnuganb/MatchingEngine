package service

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"MatchingEngine/internal/model"
	"MatchingEngine/orderBook"
)

var (
	ErrSymbolNotSpecified = errors.New("symbol not specified in order request")
	ErrChannelTimeout     = errors.New("timeout while sending order to processing channel")
)

type TradeNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderService struct {
	tradeNotifier TradeNotifier
	orderChannels map[string]chan model.OrderRequest
	mu            sync.Mutex
}

func NewOrderService(tradeNotifier TradeNotifier) *OrderService {
	return &OrderService{
		orderChannels: make(map[string]chan model.OrderRequest),
		tradeNotifier: tradeNotifier,
	}
}

func (s *OrderService) ProcessOrderRequest(req model.OrderRequest) error {
	symbol := extractSymbol(req)
	if symbol == "" {
		log.Printf("Empty symbol in order request: %+v", req)
		return ErrSymbolNotSpecified
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ch, exists := s.orderChannels[symbol]
	if !exists {
		newCh := orderBook.NewOrderBook(s.tradeNotifier)
		s.orderChannels[symbol] = newCh
		ch = newCh
		log.Printf("Created new order book and channel for symbol %s, channel addr: %p", symbol, ch)
	} else {
		log.Printf("Using existing order channel for symbol %s, channel addr: %p", symbol, ch)
	}

	select {
	case ch <- req:
		return nil
	case <-time.After(5 * time.Second):
		log.Printf("Order channel for symbol %s is full, dropping order: %+v", symbol, req)
		return ErrChannelTimeout
	}

}

func extractSymbol(req model.OrderRequest) string {
	switch req.MsgType {
	case model.MsgTypeNew:
		return req.NewOrderReq.Symbol
	case model.MsgTypeCancel:
		return req.CancelOrderReq.Symbol
	default:
		log.Printf("Invalid message type: %s", req.MsgType)
		return ""
	}
}
