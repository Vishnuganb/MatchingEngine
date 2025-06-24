package service

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"MatchingEngine/internal/model"
	"MatchingEngine/orderBook"
)

type TradeNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderService struct {
	orderBooks    map[string]*orderBook.OrderBook
	tradeNotifier TradeNotifier
	orderChannels map[string]chan model.OrderRequest
	mu            sync.Mutex
}

func NewOrderService(tradeNotifier TradeNotifier) *OrderService {
	return &OrderService{
		orderBooks:    make(map[string]*orderBook.OrderBook),
		orderChannels: make(map[string]chan model.OrderRequest),
		tradeNotifier: tradeNotifier,
	}
}

func (s *OrderService) ProcessOrderRequest(req model.OrderRequest) error {
	symbol := extractSymbol(req)
	if symbol == "" {
		log.Printf("Empty symbol in order request: %+v", req)
		return fmt.Errorf("empty symbol in order request")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ch, exists := s.orderChannels[symbol]
	if !exists {
		ob, newCh := orderBook.NewOrderBook(s.tradeNotifier)
		s.orderBooks[symbol] = ob
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
		return fmt.Errorf("channel full for symbol %s", symbol)
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
