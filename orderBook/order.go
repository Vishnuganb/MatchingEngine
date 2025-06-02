package orderBook

import "github.com/shopspring/decimal"

type Order struct {
	ID         string          `json:"id"`
	Instrument string          `json:"instrument"`
	Price      decimal.Decimal `json:"price"`
	Qty        decimal.Decimal `json:"qty"`
	LeavesQty  decimal.Decimal `json:"leaves_qty"`
	Timestamp  int64           `json:"timestamp"`
	IsBid      bool            `json:"is_bid"`
}

func (o *Order) Side() Side {
	if o.IsBid {
		return Buy
	}
	return Sell
}
