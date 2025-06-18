package orderBook

import "encoding/json"

type Trade struct {
	BuyerOrderID  string `json:"buyer_order_id"`
	SellerOrderID string `json:"seller_order_id"`
	Quantity      uint64 `json:"quantity"`
	Price         uint64 `json:"price"`
	Timestamp     int64  `json:"timestamp"`
	Instrument    string `json:"instrument"`
}

// struct to json
func (trade *Trade) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, trade)
}

func (trade *Trade) ToJSON() []byte {
	str, _ := json.Marshal(trade)
	return str
}
