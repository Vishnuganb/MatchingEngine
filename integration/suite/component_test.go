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
				`{"35": "D","new_order": {"35": "D","11": "1","54": "1","55": "BTC/USDT","38": "10","44": "100","60": 1729811234567890}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					MsgType:      "8",
					ClOrdID:      "1",
					ExecType:     model.ExecTypeNew,
					OrdStatus:    model.OrderStatusNew,
					Symbol:       "BTC/USDT",
					Side:         model.Buy,
					OrderQty:     decimal.NewFromInt(10),
					LastShares:   decimal.Zero,
					LastPx:       decimal.Zero,
					LeavesQty:    decimal.NewFromInt(10),
					CumQty:       decimal.Zero,
					AvgPx:        decimal.Zero,
					TransactTime: 1729811234567890,
				},
			},
		},
		//{
		//	name: "Matching Buy and Sell Orders",
		//	orders: []string{
		//		`{"RequestType":0,"Order":{"id":"2","side":"sell","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
		//	},
		//	expectedEvents: []interface{}{
		//		model.Trade{
		//			BuyerOrderID:  "1",
		//			SellerOrderID: "2",
		//			Quantity:      decimal.NewFromInt(10),
		//			Price:         decimal.NewFromInt(100),
		//			Instrument:    "BTC/USDT",
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "2",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(10),
		//			LeavesQty:   decimal.NewFromInt(0),
		//			CumQty:      decimal.NewFromInt(10),
		//			IsBid:       false,
		//			OrderStatus: string(orderBook.ExecTypeFill),
		//			ExecType:    string(orderBook.ExecTypeFill),
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "1",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(10),
		//			LeavesQty:   decimal.NewFromInt(0),
		//			CumQty:      decimal.NewFromInt(10),
		//			IsBid:       true,
		//			OrderStatus: string(orderBook.OrderStatusFill),
		//			ExecType:    string(orderBook.ExecTypeFill),
		//		},
		//	},
		//},
		//{
		//	name: "Matching 1 Buy and 2 Sell Orders",
		//	orders: []string{
		//		`{"RequestType":0,"Order":{"id":"3","side":"buy","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
		//		`{"RequestType":0,"Order":{"id":"4","side":"sell","qty":"5","price":"100","instrument":"BTC/USDT"}}`,
		//		`{"RequestType":0,"Order":{"id":"5","side":"sell","qty":"5","price":"100","instrument":"BTC/USDT"}}`,
		//	},
		//	expectedEvents: []interface{}{
		//		orderBook.ExecutionReport{
		//			OrderID:     "3",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(10),
		//			LeavesQty:   decimal.NewFromInt(10),
		//			CumQty:      decimal.NewFromInt(0),
		//			IsBid:       true,
		//			OrderStatus: string(orderBook.OrderStatusNew),
		//			ExecType:    string(orderBook.ExecTypeNew),
		//		},
		//		model.Trade{
		//			BuyerOrderID:  "3",
		//			SellerOrderID: "4",
		//			Quantity:      decimal.NewFromInt(5),
		//			Price:         decimal.NewFromInt(100),
		//			Instrument:    "BTC/USDT",
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "4",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(5),
		//			LeavesQty:   decimal.NewFromInt(0),
		//			CumQty:      decimal.NewFromInt(5),
		//			IsBid:       false,
		//			OrderStatus: string(orderBook.ExecTypeFill),
		//			ExecType:    string(orderBook.ExecTypeFill),
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "3",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(10),
		//			LeavesQty:   decimal.NewFromInt(5),
		//			CumQty:      decimal.NewFromInt(5),
		//			IsBid:       true,
		//			OrderStatus: string(orderBook.OrderStatusPartialFill),
		//			ExecType:    string(orderBook.ExecTypeFill),
		//		},
		//		model.Trade{
		//			BuyerOrderID:  "3",
		//			SellerOrderID: "5",
		//			Quantity:      decimal.NewFromInt(5),
		//			Price:         decimal.NewFromInt(100),
		//			Instrument:    "BTC/USDT",
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "5",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(5),
		//			LeavesQty:   decimal.NewFromInt(0),
		//			CumQty:      decimal.NewFromInt(5),
		//			IsBid:       false,
		//			OrderStatus: string(orderBook.ExecTypeFill),
		//			ExecType:    string(orderBook.ExecTypeFill),
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "3",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(10),
		//			LeavesQty:   decimal.NewFromInt(0),
		//			CumQty:      decimal.NewFromInt(10),
		//			IsBid:       true,
		//			OrderStatus: string(orderBook.OrderStatusFill),
		//			ExecType:    string(orderBook.ExecTypeFill),
		//		},
		//	},
		//},
		//{
		//	name: "Partially Matching Buy and Sell Orders",
		//	orders: []string{
		//		`{"RequestType":0,"Order":{"id":"6","side":"sell","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
		//		`{"RequestType":0,"Order":{"id":"7","side":"buy","qty":"5","price":"100","instrument":"BTC/USDT"}}`,
		//	},
		//	expectedEvents: []interface{}{
		//		orderBook.ExecutionReport{
		//			OrderID:     "6",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(10),
		//			LeavesQty:   decimal.NewFromInt(10),
		//			CumQty:      decimal.NewFromInt(0),
		//			IsBid:       false,
		//			OrderStatus: string(orderBook.OrderStatusNew),
		//			ExecType:    string(orderBook.ExecTypeNew),
		//		},
		//		model.Trade{
		//			BuyerOrderID:  "7",
		//			SellerOrderID: "6",
		//			Quantity:      decimal.NewFromInt(5),
		//			Price:         decimal.NewFromInt(100),
		//			Instrument:    "BTC/USDT",
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "7",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(5),
		//			LeavesQty:   decimal.NewFromInt(0),
		//			CumQty:      decimal.NewFromInt(5),
		//			IsBid:       true,
		//			OrderStatus: string(orderBook.OrderStatusFill),
		//			ExecType:    string(orderBook.ExecTypeFill),
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "6",
		//			Instrument:  "BTC/USDT",
		//			Price:       decimal.NewFromInt(100),
		//			OrderQty:    decimal.NewFromInt(10),
		//			LeavesQty:   decimal.NewFromInt(5),
		//			CumQty:      decimal.NewFromInt(5),
		//			IsBid:       false,
		//			OrderStatus: string(orderBook.OrderStatusPartialFill),
		//			ExecType:    string(orderBook.ExecTypeFill),
		//		},
		//	},
		//},
		//{
		//	name: "Cancel Order",
		//	orders: []string{
		//		`{"RequestType":0,"Order":{"id":"8","side":"sell","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
		//		`{"RequestType":1,"Order":{"id":"8","instrument":"BTC/USDT"}}`,
		//	},
		//	expectedEvents: []interface{}{
		//		orderBook.ExecutionReport{
		//			OrderID:     "8",
		//			Instrument:  "BTC/USDT",
		//			LeavesQty:   decimal.NewFromInt(10),
		//			CumQty:      decimal.NewFromInt(0),
		//			IsBid:       false,
		//			Price:       decimal.NewFromInt(100),
		//			OrderStatus: string(orderBook.OrderStatusNew),
		//			ExecType:    string(orderBook.ExecTypeNew),
		//		},
		//		orderBook.ExecutionReport{
		//			OrderID:     "8",
		//			Instrument:  "BTC/USDT",
		//			LeavesQty:   decimal.NewFromInt(0),
		//			IsBid:       false,
		//			Price:       decimal.NewFromInt(100),
		//			OrderStatus: string(orderBook.OrderStatusCanceled),
		//			ExecType:    string(orderBook.ExecTypeCanceled),
		//		},
		//	},
		//},
		//{
		//	name: "Reject Order",
		//	orders: []string{
		//		`{"RequestType":0,"Order":{"id":"9","side":"sell","qty":"10","price":"-100","instrument":"BTC/USDT"}}`,
		//	},
		//	expectedEvents: []interface{}{
		//		orderBook.ExecutionReport{
		//			OrderID:     "9",
		//			Instrument:  "BTC/USDT",
		//			LeavesQty:   decimal.NewFromInt(0),
		//			CumQty:      decimal.NewFromInt(0),
		//			IsBid:       false,
		//			Price:       decimal.NewFromInt(-100),
		//			OrderStatus: string(orderBook.OrderStatusRejected),
		//			ExecType:    string(orderBook.ExecTypeRejected),
		//		},
		//	},
		//},
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

					assert.Equal(t, expected.ClOrdID, actual.ClOrdID)
					assert.Equal(t, expected.Symbol, actual.Symbol)
					assert.Equal(t, expected.OrdStatus, actual.OrdStatus)
					assert.True(t, expected.LeavesQty.Equal(actual.LeavesQty), "LeavesQty mismatch")

				case model.TradeCaptureReport:
					actual, ok := receivedEvents[i].(model.TradeCaptureReport)
					require.True(t, ok, "received event is not of type model.Trade")

					assert.Equal(t, expected.Symbol, actual.Symbol)
					assert.Equal(t, expected.LastQty, actual.LastQty)
					assert.Equal(t, expected.LastPx, actual.LastPx)
					assert.Equal(t, expected.TransactTime, actual.TransactTime)

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
