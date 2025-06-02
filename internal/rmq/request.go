package rmq

import (
	"MatchingEngine/orderBook"
)

type ReqType int

const (
	ReqTypeNew ReqType = iota
	ReqTypeCancel
)

type OrderRequest struct {
	RequestType ReqType
	Order       TraderOrder
}

type TraderOrder struct {
	ID         string         `json:"id,omitempty"`
	Side       orderBook.Side `json:"side,omitempty"`
	Qty        string         `json:"qty,omitempty"`
	Price      string         `json:"price,omitempty"`
	Instrument string         `json:"instrument,omitempty"`
}
