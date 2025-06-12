package kafka

import "sync"

type MessageBatch struct {
	messages [][]byte
	mu       sync.Mutex
}

func NewMessageBatch() *MessageBatch {
	return &MessageBatch{
		messages: make([][]byte, 0),
	}
}

func (mb *MessageBatch) AddMessage(message []byte) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.messages = append(mb.messages, message)
}

func (mb *MessageBatch) GetAndClearMessages() [][]byte {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	messages := mb.messages
	mb.messages = make([][]byte, 0)
	return messages
}
