package net

type Gossip interface {
	// Broadcast a message using the gossip layer.
	// This method can block if the gossip broadcast queue is full.
	Broadcast(message Message)

	// Receive a message from the gossip layer.
	// This method blocks until a message is avaialable.
	Receive() Message

	// ReceiveQueue is a channel with received messages.
	// It enables for non-blocking message receiving.
	ReceiveQueue() <-chan Message
}
