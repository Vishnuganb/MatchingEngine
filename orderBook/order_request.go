package orderBook

import (
	"errors"

	"github.com/shopspring/decimal"
)

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

func (s Side) IsValid() bool {
	return s == Buy || s == Sell
}

type OrderRequest struct {
	ID        string
	Side      Side
	Price     decimal.Decimal
	Qty       decimal.Decimal
	Instrument string
	Timestamp int64
}

func (or *OrderRequest) Validate() error {
	switch {
	case or.ID == "":
		return errors.New("missing order id")
	case !or.Side.IsValid():
		return errors.New("invalid order side")
	case !or.Qty.IsPositive():
		return errors.New("invalid order qty")
	case or.Price.IsNegative():
		return errors.New("invalid price")
	}
	return nil
}
