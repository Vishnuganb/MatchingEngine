package service

import (
	"MatchingEngine/internal/model"
	"MatchingEngine/internal/repository"
)

type TradeService struct {
	asyncWriter repository.AsyncDBWriterInterface
}

func NewTradeService(asyncWriter repository.AsyncDBWriterInterface) *TradeService {
	return &TradeService{
		asyncWriter: asyncWriter,
	}
}

func (s *TradeService) SaveTradeAsync(trade model.Trade) {
	s.asyncWriter.EnqueueTask(repository.SaveTradeTask{
		Trade: trade,
	})
}
