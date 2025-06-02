package repository

import (
	"context"

	"github.com/shopspring/decimal"

	"MatchingEngine/orderBook"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order orderBook.Order) (orderBook.Order, error)
	SaveEvent(ctx context.Context, event orderBook.Event) (orderBook.Event, error)
	UpdateOrder(ctx context.Context, orderID string, leavesQty decimal.Decimal) (orderBook.Order, error)
	UpdateEvent(ctx context.Context, event orderBook.Event) (orderBook.Event, error)
}
