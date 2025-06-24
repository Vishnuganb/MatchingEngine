package model

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type NoSides struct {
	Side    Side   `json:"54"` // Side: 1 = Buy, 2 = Sell
	OrderID string `json:"37"`
}

// TradeCaptureReport represents a FIX AE message (Trade Capture Report)
type TradeCaptureReport struct {
	MsgType       string          `json:"35"`            // MsgType = AE (Trade Capture Report)
	TradeReportID string          `json:"571"`           // Unique ID for this Trade Report
	ExecID        string          `json:"17"`            // Execution ID
	Symbol        string          `json:"55"`            // Ticker Symbol
	LastQty       decimal.Decimal `json:"32"`            // Quantity traded
	LastPx        decimal.Decimal `json:"31"`            // Price traded
	TradeDate     string          `json:"75"`            // Format: YYYYMMDD
	TransactTime  int64           `json:"60"`            // Epoch timestamp in nanoseconds
	NoSides       []NoSides       `json:"552,omitempty"` // Repeating group for sides involved in the trade
}

// struct to json
func (trade *TradeCaptureReport) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, trade)
}

func (trade *TradeCaptureReport) ToJSON() []byte {
	str, _ := json.Marshal(trade)
	return str
}
