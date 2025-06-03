package suite

import (
	"MatchingEngine/integration/test_util"
	"encoding/json"
	"testing"
)

func TestNewOrderAddedScenario(t *testing.T) {
	// Setup RabbitMQ connection
	conn := test_util.SetupRabbitMQConnection()
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	order := []byte(`{"RequestType":0,"Order":{"id":"1","side":"buy","qty":"10","price":"100","instrument":"BTC/USDT"}}`)
	test_util.PublishOrder(ch, "order_requests", order)

	// Consume Kafka message
	message := test_util.StartConsumer("eventTopic")

	// Perform assertions
	// Unmarshal the consumed message
	var actual map[string]interface{}
	if err := json.Unmarshal([]byte(message), &actual); err != nil {
		t.Fatalf("Failed to unmarshal consumed message: %v", err)
	}

	// Define the expected fields
	expected := map[string]interface{}{
		"order_id":   "1",
		"instrument": "BTC/USDT",
		"type":       "new",
		"side":       "buy",
		"price":      "100",
		"order_qty":  "10",
		"leaves_qty": "0",
		"exec_qty":   "0",
	}

	// Compare the relevant fields
	for key, value := range expected {
		if actual[key] != value {
			t.Errorf("Expected %s to be %v, but got %v", key, value, actual[key])
		}
	}
}

func TestTradeScenario(t *testing.T) {
	conn := test_util.SetupRabbitMQConnection()
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Send sell order
	sellOrder := []byte(`{"RequestType":0,"Order":{"id":"2","side":"sell","qty":"5","price":"90","instrument":"BTC/USDT"}}`)
	test_util.PublishOrder(ch, "order_requests", sellOrder)

	// Consume Kafka message
	message := test_util.StartConsumer("eventTopic")

	// Perform assertions
	// Unmarshal the consumed message
	var actual map[string]interface{}
	if err := json.Unmarshal([]byte(message), &actual); err != nil {
		t.Fatalf("Failed to unmarshal consumed message: %v", err)
	}

	// Define the expected fields
	expected := map[string]interface{}{
		"instrument": "BTC/USDT",
		"type":       "filled",
		"side":       "buy",
		"price":      "100",
		"order_qty":  "5",
		"leaves_qty": "0",
		"exec_qty":   "5",
	}

	// Compare the relevant fields
	for key, value := range expected {
		if actual[key] != value {
			t.Errorf("Expected %s to be %v, but got %v", key, value, actual[key])
		}
	}
}

//func TestCancelOrderScenario(t *testing.T) {
//	// Setup RabbitMQ connection
//	conn := test_util.SetupRabbitMQConnection()
//	defer conn.Close()
//
//	ch, err := conn.Channel()
//	if err != nil {
//		t.Fatalf("Failed to open a channel: %v", err)
//	}
//	defer ch.Close()
//
//	// Send a new order
//	newOrder := []byte(`{"RequestType":0,"Order":{"id":"4","side":"sell","qty":"10","price":"100","instrument":"BTC/USDT"}}`)
//	test_util.PublishOrder(ch, "order_requests", newOrder)
//
//	// Send a cancel order
//	cancelOrder := []byte(`{"RequestType":1,"Order":{"id":"4"}}`)
//	test_util.PublishOrder(ch, "order_requests", cancelOrder)
//
//	// Consume Kafka message
//	message := test_util.StartConsumer()
//
//	// Perform assertions
//	// Unmarshal the consumed message
//	var actual map[string]interface{}
//	if err := json.Unmarshal([]byte(message), &actual); err != nil {
//		t.Fatalf("Failed to unmarshal consumed message: %v", err)
//	}
//
//	// Define the expected fields for the cancel event
//	expected := map[string]interface{}{
//		"order_id":   "4",
//		"instrument": "BTC/USDT",
//		"type":       "canceled",
//		"side":       "sell",
//		"price":      "100",
//		"order_qty":  "10",
//		"leaves_qty": "0",
//		"exec_qty":   "0",
//	}
//
//	// Compare the relevant fields
//	for key, value := range expected {
//		if actual[key] != value {
//			t.Errorf("Expected %s to be %v, but got %v", key, value, actual[key])
//		}
//	}
//}
