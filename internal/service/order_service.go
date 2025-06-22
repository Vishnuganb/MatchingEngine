package service

import (
	"encoding/json"
	"log"
	"sync"

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
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ch, exists := s.orderChannels[symbol]
	if !exists {
		ob, ch := orderBook.NewOrderBook(s.tradeNotifier)
		s.orderBooks[symbol] = ob
		s.orderChannels[symbol] = ch
	}

	select {
	case ch <- req:
		return nil
	default:
		return nil
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
