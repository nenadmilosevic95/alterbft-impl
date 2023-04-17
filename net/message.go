package net

import "crypto/sha256"

// Message is a generic message sent via network.
type Message []byte

// MessageID is a message identifier.
type MessageID [sha256.Size]byte

// ID returns an unique identifier for the message.
func (m Message) ID() MessageID {
	return sha256.Sum256(m)
}

// Code returns the message code, an indication of its type.
func (m Message) Code() byte {
	return m[0]
}
