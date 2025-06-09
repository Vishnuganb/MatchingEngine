package orderBook

import "encoding/json"

type MockKafkaProducer struct{}

func (m *MockKafkaProducer) NotifyEventAndTrade(orderID string, value json.RawMessage) error {
	// Simulate a no-op or log for test purposes
	return nil
}