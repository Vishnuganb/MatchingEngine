package service

import (
	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
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

func (s *OrderService) UpdateOrderAsync(orderID, orderStatus string, leavesQty, cumQty, price decimal.Decimal) {
	task := repository.UpdateOrderTask{
		OrderID:     orderID,
		OrderStatus: orderStatus,
		LeavesQty:   leavesQty,
		CumQty:      cumQty,
		Price:       price,
	}
	s.asyncWriter.EnqueueTask(task)
}
