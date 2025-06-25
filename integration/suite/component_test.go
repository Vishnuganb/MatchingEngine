//go:build integration

package suite

import (
	"encoding/json"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"MatchingEngine/integration/test_util"
	"MatchingEngine/internal/model"
)

var (
	once   sync.Once
	Events <-chan string
)

func setupConsumerOnce() {
	once.Do(func() {
		topic := "executionTopic"
		Events = test_util.ConsumeKafkaMessages(topic)
	})
}

func TestOrderFlowScenarios(t *testing.T) {
	setupConsumerOnce()
	tests := []struct {
		name           string
		orders         []string
		expectedEvents []interface{}
	}{
		{
			name: "New Buy Order",
			orders: []string{
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-001","54": "1","55": "BTC/USDT","38": "10","44": "100","60": 1729811234567890}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-001",
					ExecType:     model.ExecTypeNew,
					OrdStatus:    model.OrderStatusNew,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(0),
					LastPx:       decimal.NewFromInt(0),
					LeavesQty:    decimal.NewFromInt(10),
					CumQty:       decimal.NewFromInt(0),
					AvgPx:        decimal.NewFromInt(0),
					TransactTime: 1729811234567890,
				},
			},
		},
		{
			name: "Matching Buy and Sell Orders",
			orders: []string{
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-002","54": "2","55": "BTC/USDT","38": "10","44": "100","60": 1729811234567890}}`,
			},
			expectedEvents: []interface{}{
				model.TradeCaptureReport{
					MsgType: "AE",
					Symbol:  "BTC/USDT",
					LastQty: decimal.NewFromInt(10),
					LastPx:  decimal.NewFromInt(100),
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-002",
					ExecType:     model.ExecTypeFill,
					OrdStatus:    model.OrderStatusFill,
					Symbol:       "BTC/USDT",
					Side:         model.Sell,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(10),
					LastPx:       decimal.NewFromInt(100),
					LeavesQty:    decimal.NewFromInt(0),
					CumQty:       decimal.NewFromInt(10),
					AvgPx:        decimal.NewFromInt(0),
					TransactTime: 1729811234567890,
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-001",
					ExecType:     model.ExecTypeFill,
					OrdStatus:    model.OrderStatusFill,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(10),
					LastPx:       decimal.NewFromInt(100),
					LeavesQty:    decimal.NewFromInt(0),
					CumQty:       decimal.NewFromInt(10),
					AvgPx:        decimal.NewFromInt(100),
					TransactTime: 1729811234567890,
				},
			},
		},
		{
			name: "Matching 1 Buy and 2 Sell Orders",
			orders: []string{
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-003","54": "1","55": "BTC/USDT","38": "10","44": "100","60": 1729811234567890}}`,
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-004","54": "2","55": "BTC/USDT","38": "5","44": "100","60": 1729811234567890}}`,
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-005","54": "2","55": "BTC/USDT","38": "5","44": "100","60": 1729811234567890}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-003",
					ExecType:     model.ExecTypeNew,
					OrdStatus:    model.OrderStatusNew,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(0),
					LastPx:       decimal.NewFromInt(0),
					LeavesQty:    decimal.NewFromInt(10),
					CumQty:       decimal.NewFromInt(0),
					AvgPx:        decimal.NewFromInt(0),
					TransactTime: 1729811234567890,
				},
				model.TradeCaptureReport{
					MsgType: "AE",
					Symbol:  "BTC/USDT",
					LastQty: decimal.NewFromInt(5),
					LastPx:  decimal.NewFromInt(100),
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-004",
					ExecType:     model.ExecTypeFill,
					OrdStatus:    model.OrderStatusFill,
					Symbol:       "BTC/USDT",
					Side:         model.Sell,
					OrderQty:     decimal.NewFromInt(5),
					LastShares:   decimal.NewFromInt(5),
					LastPx:       decimal.NewFromInt(100),
					LeavesQty:    decimal.NewFromInt(0),
					CumQty:       decimal.NewFromInt(5),
					AvgPx:        decimal.NewFromInt(100),
					TransactTime: 1729811234567890,
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-003",
					ExecType:     model.ExecTypeFill,
					OrdStatus:    model.OrderStatusPartialFill,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(5),
					LastPx:       decimal.NewFromInt(100),
					LeavesQty:    decimal.NewFromInt(5),
					CumQty:       decimal.NewFromInt(5),
					AvgPx:        decimal.NewFromInt(100),
					TransactTime: 1729811234567890,
				},
				model.TradeCaptureReport{
					MsgType: "AE",
					Symbol:  "BTC/USDT",
					LastQty: decimal.NewFromInt(5),
					LastPx:  decimal.NewFromInt(100),
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-005",
					ExecType:     model.ExecTypeFill,
					OrdStatus:    model.OrderStatusFill,
					Symbol:       "BTC/USDT",
					Side:         model.Sell,
					OrderQty:     decimal.NewFromInt(5),
					LastShares:   decimal.NewFromInt(5),
					LastPx:       decimal.NewFromInt(100),
					LeavesQty:    decimal.NewFromInt(0),
					CumQty:       decimal.NewFromInt(5),
					AvgPx:        decimal.NewFromInt(100),
					TransactTime: 1729811234567890,
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-003",
					ExecType:     model.ExecTypeFill,
					OrdStatus:    model.OrderStatusFill,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(5),
					LastPx:       decimal.NewFromInt(100),
					LeavesQty:    decimal.NewFromInt(0),
					CumQty:       decimal.NewFromInt(10),
					AvgPx:        decimal.NewFromInt(100),
					TransactTime: 1729811234567890,
				},
			},
		},
		{
			name: "Partially Matching Buy and Sell Orders",
			orders: []string{
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-006","54": "1","55": "BTC/USDT","38": "10","44": "100","60": 1729811234567890}}`,
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-007","54": "2","55": "BTC/USDT","38": "5","44": "100","60": 1729811234567890}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-006",
					ExecType:     model.ExecTypeNew,
					OrdStatus:    model.OrderStatusNew,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(0),
					LastPx:       decimal.NewFromInt(0),
					LeavesQty:    decimal.NewFromInt(10),
					CumQty:       decimal.NewFromInt(0),
					AvgPx:        decimal.NewFromInt(0),
					TransactTime: 1729811234567890,
				},
				model.TradeCaptureReport{
					MsgType: "AE",
					Symbol:  "BTC/USDT",
					LastQty: decimal.NewFromInt(5),
					LastPx:  decimal.NewFromInt(100),
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-007",
					ExecType:     model.ExecTypeFill,
					OrdStatus:    model.OrderStatusFill,
					Symbol:       "BTC/USDT",
					Side:         model.Sell,
					OrderQty:     decimal.NewFromInt(5),
					LastShares:   decimal.NewFromInt(5),
					LastPx:       decimal.NewFromInt(100),
					LeavesQty:    decimal.NewFromInt(0),
					CumQty:       decimal.NewFromInt(5),
					AvgPx:        decimal.NewFromInt(0),
					TransactTime: 1729811234567890,
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-006",
					ExecType:     model.ExecTypeFill,
					OrdStatus:    model.OrderStatusPartialFill,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(5),
					LastPx:       decimal.NewFromInt(100),
					LeavesQty:    decimal.NewFromInt(5),
					CumQty:       decimal.NewFromInt(5),
					AvgPx:        decimal.NewFromInt(100),
					TransactTime: 1729811234567890,
				},
			},
		},
		{
			name: "Cancel Order",
			orders: []string{
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-008","54": "1","55": "BTC/USDT","38": "10","44": "100","60": 1729811234567890}}`,
				`{"35": "F","cancel_order": {"35": "F","11": "clOrdId-009","54": "1","55": "BTC/USDT","38": "10","44": "100","60": 1729811234567890,"41": "clOrdId-008"}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-008",
					ExecType:     model.ExecTypeNew,
					OrdStatus:    model.OrderStatusNew,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(0),
					LastPx:       decimal.NewFromInt(0),
					LeavesQty:    decimal.NewFromInt(10),
					CumQty:       decimal.NewFromInt(0),
					AvgPx:        decimal.NewFromInt(0),
					TransactTime: 1729811234567890,
				},
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-008",
					ExecType:     model.ExecTypeCanceled,
					OrdStatus:    model.OrderStatusCanceled,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(0),
					LastPx:       decimal.NewFromInt(0),
					LeavesQty:    decimal.NewFromInt(0),
					CumQty:       decimal.NewFromInt(0),
					AvgPx:        decimal.NewFromInt(0),
					TransactTime: 1729811234567890,
				},
			},
		},
		{
			name: "Reject Order",
			orders: []string{
				`{"35": "D","new_order": {"35": "D","11": "clOrdId-010","54": "1","55": "BTC/USDT","38": "10","44": "-100","60": 1729811234567890}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "clOrdId-010",
					ExecType:     model.ExecTypeRejected,
					OrdStatus:    model.OrderStatusRejected,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.NewFromInt(0),
					LastPx:       decimal.NewFromInt(0),
					LeavesQty:    decimal.NewFromInt(0),
					CumQty:       decimal.NewFromInt(0),
					AvgPx:        decimal.NewFromInt(0),
					TransactTime: 1729811234567890,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			conn := test_util.SetupRabbitMQConnection()
			defer conn.Close()

			ch, err := conn.Channel()
			require.NoError(t, err)
			defer ch.Close()

			// Clean up before test
			_, err = ch.QueuePurge("orderRequests", false)
			if err != nil {
				t.Fatalf("Failed to purge RabbitMQ queue: %v", err)
			}

			// Send orders
			for i, order := range tt.orders {
				log.Printf("Publishing order %d: %s", i+1, order)
				test_util.PublishOrder(ch, "orderRequests", []byte(order))
				if i < len(tt.orders)-1 {
					log.Printf("Waiting before sending next order...")
					time.Sleep(5 * time.Second) // Added a small delay between orders
				}
			}

			log.Printf("All orders published, waiting for events...")

			// Collect events
			var receivedEvents []interface{}
			timeout := time.After(3 * time.Minute)
			expectedCount := len(tt.expectedEvents)

			startTime := time.Now()
			for len(receivedEvents) < expectedCount {
				select {
				case message, ok := <-Events:
					if !ok {
						// Channel closed, check if we got all expected events
						if len(receivedEvents) < expectedCount {
							t.Fatalf("Message channel closed before receiving all events. Got %d/%d events",
								len(receivedEvents), expectedCount)
						}
						break
					}

					event := parseKafkaEvent(message)
					if matchesExpectedEvent(event, tt.expectedEvents) {
						receivedEvents = append(receivedEvents, event)
					}

				case <-timeout:
					t.Fatalf("Timeout waiting for events. Got %d/%d events after %v",
						len(receivedEvents), expectedCount, time.Since(startTime))
				}
			}

			assert.Equal(t, len(tt.expectedEvents), len(receivedEvents))
			for _, event := range receivedEvents {
				log.Printf("Received event: %+v", event)
			}
			for i, exp := range tt.expectedEvents {
				switch expected := exp.(type) {
				case model.ExecutionReport:
					actual, ok := receivedEvents[i].(model.ExecutionReport)
					require.True(t, ok, "received event is not of type model.ExecutionReport")

					assert.Equal(t, expected.MsgType, actual.MsgType)
					assert.Equal(t, expected.ExecType, actual.ExecType)
					assert.Equal(t, expected.Side, actual.Side)
					assert.Equal(t, expected.OrderQty, actual.OrderQty)
					assert.Equal(t, expected.LastShares, actual.LastShares)
					assert.Equal(t, expected.LastPx, actual.LastPx)
					assert.Equal(t, expected.CumQty, actual.CumQty)
					assert.Equal(t, expected.ClOrdID, actual.ClOrdID)
					assert.Equal(t, expected.Symbol, actual.Symbol)
					assert.Equal(t, expected.OrdStatus, actual.OrdStatus)
					assert.True(t, expected.LeavesQty.Equal(actual.LeavesQty), "LeavesQty mismatch")

				case model.TradeCaptureReport:
					actual, ok := receivedEvents[i].(model.TradeCaptureReport)
					require.True(t, ok, "received event is not of type model.Trade")

					assert.Equal(t, expected.MsgType, actual.MsgType)
					assert.Equal(t, expected.Symbol, actual.Symbol)
					assert.Equal(t, expected.LastQty, actual.LastQty)
					assert.Equal(t, expected.LastPx, actual.LastPx)

				default:
					t.Fatalf("unexpected event type: %T", exp)
				}
			}
		})
	}
}

func parseKafkaEvent(msg string) interface{} {
	var raw map[string]interface{}
	_ = json.Unmarshal([]byte(msg), &raw)

	if _, ok := raw["571"]; ok {
		var trade model.TradeCaptureReport
		_ = json.Unmarshal([]byte(msg), &trade)
		return trade
	}

	var report model.ExecutionReport
	_ = json.Unmarshal([]byte(msg), &report)
	return report
}

func matchesExpectedEvent(event interface{}, expectedList []interface{}) bool {
	switch evt := event.(type) {
	case model.ExecutionReport:
		for _, e := range expectedList {
			if exp, ok := e.(model.ExecutionReport); ok && exp.ClOrdID == evt.ClOrdID {
				return true
			}
		}
	case model.TradeCaptureReport:
		for _, e := range expectedList {
			if _, ok := e.(model.TradeCaptureReport); ok {
				return true
			}
		}
	}
	return false
}
