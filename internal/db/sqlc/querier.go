// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"context"
)

type Querier interface {
	CreateActiveOrder(ctx context.Context, arg CreateActiveOrderParams) (ActiveOrder, error)
	DeleteActiveOrder(ctx context.Context, id string) (ActiveOrder, error)
	GetActiveOrder(ctx context.Context, id string) (ActiveOrder, error)
	ListActiveOrders(ctx context.Context) ([]ActiveOrder, error)
	UpdateActiveOrder(ctx context.Context, arg UpdateActiveOrderParams) (ActiveOrder, error)
}

var _ Querier = (*Queries)(nil)
