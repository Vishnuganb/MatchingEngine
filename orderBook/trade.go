package orderBook

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Trade struct {
	BuyerOrderID  string          `json:"buyer_order_id"`
	SellerOrderID string          `json:"seller_order_id"`
	Quantity      decimal.Decimal `json:"quantity"`
	Price         decimal.Decimal `json:"price"`
	Timestamp     int64           `json:"timestamp"`
	Instrument    string          `json:"instrument"`
}

// struct to json
func (trade *Trade) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, trade)
}

func (trade *Trade) ToJSON() []byte {
	str, _ := json.Marshal(trade)
	return str
}
