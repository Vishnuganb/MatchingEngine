package service

import (
	"MatchingEngine/internal/model"
	"MatchingEngine/internal/repository"
)

type TradeService struct {
	asyncWriter repository.AsyncDBWriterInterface
	Notifier Notifier
}

func NewTradeService(asyncWriter repository.AsyncDBWriterInterface, notifier Notifier) *TradeService {
	return &TradeService{
		asyncWriter: asyncWriter,
		Notifier: notifier,
	}
}

func (s *TradeService) SaveTradeAsync(trade model.TradeCaptureReport) {
	s.asyncWriter.EnqueueTask(repository.SaveTradeTask{
		Trade: trade,
	})
}
