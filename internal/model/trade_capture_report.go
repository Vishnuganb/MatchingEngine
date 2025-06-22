package model

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// Party represents a FIX party (buyer, seller, broker, etc.)
type Party struct {
	PartyID   string `json:"448"` // Party identifier (e.g., firm ID, trader ID)
	PartyRole int    `json:"452"` // FIX-defined role (e.g., 1=Executing Firm, 17=Contra Firm)
}

// TradeCaptureReport represents a FIX AE message (Trade Capture Report)
type TradeCaptureReport struct {
	MsgType            string          `json:"35"`            // MsgType = AE (Trade Capture Report)
	TradeReportID      string          `json:"571"`           // Unique ID for this Trade Report
	ExecID             string          `json:"17"`            // Execution ID
	OrderID            string          `json:"37"`            // OrderID (associated with one side of the trade)
	ClOrdID            string          `json:"11,omitempty"`  // Client Order ID
	Symbol             string          `json:"55"`            // Ticker Symbol
	Side               Side            `json:"54"`            // Side: 1 = Buy, 2 = Sell
	LastQty            decimal.Decimal `json:"32"`            // Quantity traded
	LastPx             decimal.Decimal `json:"31"`            // Price traded
	TradeDate          string          `json:"75"`            // Format: YYYYMMDD
	TransactTime       int64           `json:"60"`            // Epoch timestamp in nanoseconds
	PreviouslyReported bool            `json:"570"`           // true = already reported to counterparty
	Text               string          `json:"58,omitempty"`  // Optional free-text note
	Parties            []Party         `json:"453,omitempty"` // Repeating group for party identification
}

// struct to json
func (trade *TradeCaptureReport) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, trade)
}

func (trade *TradeCaptureReport) ToJSON() []byte {
	str, _ := json.Marshal(trade)
	return str
}
