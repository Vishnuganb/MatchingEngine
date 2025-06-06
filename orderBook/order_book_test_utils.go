package orderBook

import "encoding/json"

type MockKafkaProducer struct{}

func (m *MockKafkaProducer) NotifyEventAndOrder(orderID string, value json.RawMessage) error {
	// Simulate a no-op or log for test
	return nil
}

