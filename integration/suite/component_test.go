//go:build integration

package suite

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"MatchingEngine/integration/test_util"
	"MatchingEngine/internal/model"
	"MatchingEngine/orderBook"
)

func TestOrderFlowScenarios(t *testing.T) {
	tests := []struct {
		name           string
		orders         []string
		expectedEvents []interface{}
	}{
		{
			name: "New Buy Order",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"1","side":"buy","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					OrderID:     "1",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(10),
					CumQty:      decimal.NewFromInt(0),
					IsBid:       true,
					OrderStatus: string(orderBook.OrderStatusNew),
					ExecType:    string(orderBook.ExecTypeNew),
				},
			},
		},
		{
			name: "Matching Buy and Sell Orders",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"2","side":"sell","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []interface{}{
				model.Trade{
					BuyerOrderID:  "1",
					SellerOrderID: "2",
					Quantity:      10,
					Price:         100,
				},
				model.ExecutionReport{
					OrderID:     "2",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(0),
					CumQty:      decimal.NewFromInt(10),
					IsBid:       false,
					OrderStatus: string(orderBook.ExecTypeFill),
					ExecType:    string(orderBook.ExecTypeFill),
				},
				model.ExecutionReport{
					OrderID:     "1",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(0),
					CumQty:      decimal.NewFromInt(10),
					IsBid:       true,
					OrderStatus: string(orderBook.OrderStatusFill),
					ExecType:    string(orderBook.ExecTypeFill),
				},
			},
		},
		{
			name: "Matching 1 Buy and 2 Sell Orders",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"3","side":"buy","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
				`{"RequestType":0,"Order":{"id":"4","side":"sell","qty":"5","price":"100","instrument":"BTC/USDT"}}`,
				`{"RequestType":0,"Order":{"id":"5","side":"sell","qty":"5","price":"100","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []interface{}{
				model.Trade{
					BuyerOrderID:  "3",
					SellerOrderID: "4",
					Quantity:      5,
					Price:         100,
				},
				model.ExecutionReport{
					OrderID:     "4",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(5),
					LeavesQty:   decimal.NewFromInt(0),
					CumQty:      decimal.NewFromInt(5),
					IsBid:       false,
					OrderStatus: string(orderBook.ExecTypeFill),
					ExecType:    string(orderBook.ExecTypeFill),
				},
				model.ExecutionReport{
					OrderID:     "3",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(5),
					CumQty:      decimal.NewFromInt(5),
					IsBid:       true,
					OrderStatus: string(orderBook.OrderStatusPartialFill),
					ExecType:    string(orderBook.ExecTypeFill),
				},
				model.Trade{
					BuyerOrderID:  "3",
					SellerOrderID: "5",
					Quantity:      5,
					Price:         100,
				},
				model.ExecutionReport{
					OrderID:     "5",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(5),
					LeavesQty:   decimal.NewFromInt(0),
					CumQty:      decimal.NewFromInt(5),
					IsBid:       false,
					OrderStatus: string(orderBook.ExecTypeFill),
					ExecType:    string(orderBook.ExecTypeFill),
				},
				model.ExecutionReport{
					OrderID:     "5",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(0),
					CumQty:      decimal.NewFromInt(10),
					IsBid:       true,
					OrderStatus: string(orderBook.OrderStatusFill),
					ExecType:    string(orderBook.ExecTypeFill),
				},
			},
		},
		{
			name: "Partially Matching Buy and Sell Orders",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"6","side":"sell","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
				`{"RequestType":0,"Order":{"id":"7","side":"buy","qty":"5","price":"100","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					OrderID:     "6",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(10),
					CumQty:      decimal.NewFromInt(0),
					IsBid:       false,
					OrderStatus: string(orderBook.OrderStatusNew),
					ExecType:    string(orderBook.ExecTypeNew),
				},
				model.Trade{
					BuyerOrderID:  "7",
					SellerOrderID: "6",
					Quantity:      5,
					Price:         100,
				},
				model.ExecutionReport{
					OrderID:     "7",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(5),
					LeavesQty:   decimal.NewFromInt(0),
					CumQty:      decimal.NewFromInt(5),
					IsBid:       true,
					OrderStatus: string(orderBook.OrderStatusFill),
					ExecType:    string(orderBook.ExecTypeFill),
				},
				model.ExecutionReport{
					OrderID:     "6",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					OrderQty:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(5),
					CumQty:      decimal.NewFromInt(5),
					IsBid:       false,
					OrderStatus: string(orderBook.OrderStatusPartialFill),
					ExecType:    string(orderBook.ExecTypeFill),
				},
			},
		},
		{
			name: "Cancel Order",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"8","side":"sell","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
				`{"RequestType":1,"Order":{"id":"8","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					OrderID:     "8",
					Instrument:  "BTC/USDT",
					LeavesQty:   decimal.NewFromInt(10),
					IsBid:       false,
					Price:       decimal.NewFromInt(100),
					OrderStatus: string(orderBook.OrderStatusNew),
					ExecType:    string(orderBook.ExecTypeNew),
				},
				model.ExecutionReport{
					OrderID:     "8",
					Instrument:  "BTC/USDT",
					LeavesQty:   decimal.NewFromInt(0),
					IsBid:       false,
					Price:       decimal.NewFromInt(100),
					OrderStatus: string(orderBook.OrderStatusCanceled),
					ExecType:    string(orderBook.ExecTypeCanceled),
				},
			},
		},
		{
			name: "Reject Order",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"9","side":"sell","qty":"10","price":"-100","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []interface{}{
				model.ExecutionReport{
					OrderID:     "9",
					Instrument:  "BTC/USDT",
					LeavesQty:   decimal.NewFromInt(10),
					IsBid:       false,
					Price:       decimal.NewFromInt(-100),
					OrderStatus: string(orderBook.OrderStatusRejected),
					ExecType:    string(orderBook.ExecTypeRejected),
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
			messageChan := test_util.ConsumeKafkaMessages("executionTopic")
			var receivedEvents []interface{}
			timeout := time.After(2 * time.Minute)
			expectedCount := len(tt.expectedEvents)

			startTime := time.Now()
			for len(receivedEvents) < expectedCount {
				select {
				case message, ok := <-messageChan:
					if !ok {
						// Channel closed, check if we got all expected events
						if len(receivedEvents) < expectedCount {
							t.Fatalf("Message channel closed before receiving all events. Got %d/%d events",
								len(receivedEvents), expectedCount)
						}
						break
					}

					var raw map[string]interface{}
					if err := json.Unmarshal([]byte(message), &raw); err != nil {
						log.Printf("Failed to unmarshal message into map: %v", err)
						continue
					}

					if _, isTrade := raw["buyer_order_id"]; !isTrade {
						var event model.ExecutionReport
						if err := json.Unmarshal([]byte(message), &event); err != nil {
							log.Printf("Failed to unmarshal ExecutionReport: %v", err)
							continue
						}
						receivedEvents = append(receivedEvents, event)
					} else {
						var trade model.Trade
						if err := json.Unmarshal([]byte(message), &trade); err != nil {
							log.Printf("Failed to unmarshal Trade: %v", err)
							continue
						}
						receivedEvents = append(receivedEvents, trade)
					}

					// Break if we've received all expected events
					if len(receivedEvents) == expectedCount {
						break
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

					assert.Equal(t, expected.OrderID, actual.OrderID)
					assert.Equal(t, expected.Instrument, actual.Instrument)
					assert.Equal(t, expected.OrderStatus, actual.OrderStatus)
					assert.True(t, expected.LeavesQty.Equal(actual.LeavesQty), "LeavesQty mismatch")

				case model.Trade:
					actual, ok := receivedEvents[i].(model.Trade)
					require.True(t, ok, "received event is not of type model.Trade")

					assert.Equal(t, expected.BuyerOrderID, actual.BuyerOrderID)
					assert.Equal(t, expected.SellerOrderID, actual.SellerOrderID)
					assert.Equal(t, expected.Quantity, actual.Quantity)
					assert.Equal(t, expected.Price, actual.Price)

				default:
					t.Fatalf("unexpected event type: %T", exp)
				}
			}
		})
	}
}
