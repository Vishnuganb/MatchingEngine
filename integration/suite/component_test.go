package suite

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"log"
	"testing"
	"time"

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
		expectedEvents []model.OrderEvent
	}{
		{
			name: "New Buy Order",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"1","side":"buy","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []model.OrderEvent{
				{
					EventType:   string(orderBook.EventTypeNew),
					OrderID:     "1",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					Quantity:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(10),
					ExecQty:     decimal.NewFromInt(0),
					IsBid:       true,
					OrderStatus: string(orderBook.EventTypeNew),
				},
			},
		},
		{
			name: "Matching Buy and Sell Orders",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"1","side":"buy","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
				`{"RequestType":0,"Order":{"id":"2","side":"sell","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []model.OrderEvent{
				{
					EventType:   string(orderBook.EventTypeFill),
					OrderID:     "1",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					Quantity:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(0),
					ExecQty:     decimal.NewFromInt(10),
					IsBid:       true,
					OrderStatus: string(orderBook.EventTypeFill),
				},
				{
					EventType:   string(orderBook.EventTypeFill),
					OrderID:     "2",
					Instrument:  "BTC/USDT",
					Price:       decimal.NewFromInt(100),
					Quantity:    decimal.NewFromInt(10),
					LeavesQty:   decimal.NewFromInt(0),
					ExecQty:     decimal.NewFromInt(10),
					IsBid:       false,
					OrderStatus: string(orderBook.EventTypeFill),
				},
			},
		},
		{
			name: "Cancel Order",
			orders: []string{
				`{"RequestType":0,"Order":{"id":"1","side":"buy","qty":"10","price":"100","instrument":"BTC/USDT"}}`,
				`{"RequestType":1,"Order":{"id":"1","instrument":"BTC/USDT"}}`,
			},
			expectedEvents: []model.OrderEvent{
				{
					EventType:   string(orderBook.EventTypeNew),
					OrderID:     "1",
					Instrument:  "BTC/USDT",
					LeavesQty:   decimal.NewFromInt(10),
					OrderStatus: string(orderBook.EventTypeNew),
				},
				{
					EventType:   string(orderBook.EventTypeCanceled),
					OrderID:     "1",
					Instrument:  "BTC/USDT",
					LeavesQty:   decimal.NewFromInt(0),
					OrderStatus: string(orderBook.EventTypeCanceled),
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

			test_util.ClearKafkaTopic("eventTopic")

			// Clean up before test
			_, err = ch.QueuePurge("order_requests", false)
			if err != nil {
				t.Fatalf("Failed to purge RabbitMQ queue: %v", err)
			}

			// Send orders
			for i, order := range tt.orders {
				log.Printf("Publishing order %d: %s", i+1, order)
				test_util.PublishOrder(ch, "order_requests", []byte(order))
				if i < len(tt.orders)-1 {
					log.Printf("Waiting before sending next order...")
					time.Sleep(10 * time.Second) // Added a small delay between orders
				}
			}

			log.Printf("All orders published, waiting for events...")

			// Collect events
			var receivedEvents []model.OrderEvent
			timeout := time.After(1 * time.Minute)
			expectedCount := len(tt.expectedEvents)

			startTime := time.Now()
			for len(receivedEvents) < expectedCount {
				select {
				case message := <-test_util.ConsumeKafkaMessages("eventTopic"):
					var event model.OrderEvent
					err := json.Unmarshal([]byte(message), &event)
					require.NoError(t, err)
					log.Printf("[%v] Received event: Type=%s, OrderID=%s, Status=%s",
						time.Since(startTime), event.EventType, event.OrderID, event.OrderStatus)
					receivedEvents = append(receivedEvents, event)
				case <-timeout:
					log.Printf("Timeout after %v. Got %d/%d events",
						time.Since(startTime), len(receivedEvents), expectedCount)
					t.Fatalf("Timeout waiting for events. Received events %d / Expected events %d",
						len(receivedEvents), expectedCount)
				}
			}

			// Verify events
			assert.Equal(t, len(tt.expectedEvents), len(receivedEvents))
			for i, expected := range tt.expectedEvents {
				actual := receivedEvents[i]
				assert.Equal(t, expected.EventType, actual.EventType)
				assert.Equal(t, expected.OrderID, actual.OrderID)
				assert.Equal(t, expected.Instrument, actual.Instrument)
				assert.Equal(t, expected.OrderStatus, actual.OrderStatus)
				assert.Equal(t, expected.LeavesQty, actual.LeavesQty)
			}
		})
	}
}

func TestInvalidOrderScenarios(t *testing.T) {
	tests := []struct {
		name          string
		order         string
		expectedError string
	}{
		{
			name:          "Invalid JSON",
			order:         `{"invalid json"`,
			expectedError: "invalid message format",
		},
		{
			name:          "Missing Required Fields",
			order:         `{"RequestType":0,"Order":{"id":"1"}}`,
			expectedError: "missing required fields",
		},
		{
			name:          "Invalid Price Format",
			order:         `{"RequestType":0,"Order":{"id":"1","side":"buy","qty":"10","price":"invalid","instrument":"BTC/USDT"}}`,
			expectedError: "invalid price format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := test_util.SetupRabbitMQConnection()
			defer conn.Close()

			ch, err := conn.Channel()
			require.NoError(t, err)
			defer ch.Close()

			test_util.PublishOrder(ch, "order_requests", []byte(tt.order))

			// Wait for error event
            message := <-test_util.ConsumeKafkaMessages("eventTopic")
            var event model.OrderEvent
            err = json.Unmarshal([]byte(message), &event)
			require.NoError(t, err)

			assert.Equal(t, string(orderBook.EventTypeRejected), event.EventType)
			assert.Contains(t, event.OrderStatus, tt.expectedError)
		})
	}
}
