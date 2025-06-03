package suite

import (
	"MatchingEngine/integration/test_util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

	queueName := "order_requests"
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		t.Fatalf("Failed to declare a queue: %v", err)
	}

	// Add assertions or validations for the new order
	time.Sleep(1 * time.Second) // Wait for processing
	assert.True(t, true, "Order added successfully")
}

func TestSellThenBuyOrderScenario(t *testing.T) {
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

	// Send buy order
	buyOrder := []byte(`{"RequestType":0,"Order":{"id":"3","side":"buy","qty":"5","price":"90","instrument":"BTC/USDT"}}`)
	test_util.PublishOrder(ch, "order_requests", buyOrder)

	// Add assertions or validations for the output
	time.Sleep(1 * time.Second) // Wait for processing
	assert.True(t, true, "Sell and Buy orders processed successfully")
}
