package service

import (
	"MatchingEngine/internal/model"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/repository"
)

type OrderService struct {
	asyncWriter repository.AsyncDBWriterInterface
}

func NewOrderService(asyncWriter repository.AsyncDBWriterInterface) *OrderService {
	return &OrderService{
		asyncWriter: asyncWriter,
	}
}

func (s *OrderService) SaveOrderAsync(order model.Order) {
	s.asyncWriter.EnqueueTask(repository.SaveOrderTask{
		Order: order,
	})
}

func (s *OrderService) SaveEventAsync(event model.Event) {
	s.asyncWriter.EnqueueTask(repository.SaveEventTask{Event: event})
}

func (s *OrderService) UpdateOrderAsync(orderID string, leavesQty decimal.Decimal) {
	task := repository.UpdateOrderTask{
		OrderID:   orderID,
		LeavesQty: leavesQty,
	}
	s.asyncWriter.EnqueueTask(task)
}

func (s *OrderService) CancelEventAsync(event model.Event) {
	task := repository.CancelEventTask{
		Event: event,
	}
	s.asyncWriter.EnqueueTask(task)
}
