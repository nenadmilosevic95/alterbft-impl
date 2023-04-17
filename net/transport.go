package net

type Transport interface {
	// Broadcast a message using the transport.
	// This method can block if the gossip broadcast queue is full.
	Broadcast(message Message)

	// Receive a message from the transport.
	// This method blocks until a message is avaialable.
	Receive() Message

	// ReceiveQueue is a channel with received messages.
	// It enables for non-blocking message receiving.
	ReceiveQueue() <-chan Message

	// Send a message to a set of destinations.
	// The process ID id of the destinations should be provided.
	Send(message Message, pids ...int)
}
