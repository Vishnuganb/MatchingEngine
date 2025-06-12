package suite

import (
	"encoding/json"
	"testing"

	"MatchingEngine/integration/test_util"
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

	// Purge RabbitMQ queue
	_, err = ch.QueuePurge("order_requests", false)
	if err != nil {
		t.Fatalf("Failed to purge RabbitMQ queue: %v", err)
	}

	// Consume and discard all Kafka messages
	test_util.ClearKafkaTopic("eventTopic")

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
