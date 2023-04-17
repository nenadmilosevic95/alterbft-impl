package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// Value proposed for consensus.
type Value []byte

// Equal returns whether a value equals another.
func (v Value) Equal(vv Value) bool {
	return bytes.Equal(v, vv)
}

// ID returns a hash uniquely identifying the value.
// The value ID is a value itself, but smaller than a full value.
func (v Value) ID() Value {
	if v == nil {
		return nil
	}
	h := sha256.Sum256(v)
	return h[:]
}

// Key produces a fixed-length key for use in indexing.
func (v Value) Key() ValueKey {
	return sha256.Sum256(v)
}

// String returns the hex-encoded value as a string.
func (v Value) String() string {
	switch len(v) {
	case 0:
		return "nil"
	case sha256.Size: // ID
		return fmt.Sprintf("%X", []byte(v[:8]))
	default:
		return fmt.Sprintf("%X(%d)", []byte(v.ID()[:8]), len(v))
	}
}

// ValueKey is a fixed length array key used as a value index.
type ValueKey [sha256.Size]byte
