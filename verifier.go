package tendermint

import (
	"log"
	"sync"

	"dslab.inf.usi.ch/tendermint/bootstrap"
	"dslab.inf.usi.ch/tendermint/consensus"
	"dslab.inf.usi.ch/tendermint/crypto"
	"dslab.inf.usi.ch/tendermint/net"
	"github.com/hashicorp/golang-lru/simplelru"
)

// VerifierChanSize is the capacity of input and output channels.
var VerifierChanSize = 64

// VerifierCacheSize is the size of the cached of verified signatures.
var VerifierCacheSize = 1024

// Verifier unmarshalls and verifies signatures of consensus messages.
type Verifier struct {
	keys    []crypto.PublicKey
	input   <-chan net.Message
	output  chan *consensus.Message
	skipped chan net.Message

	cache   simplelru.LRUCache
	stats   VerifierStats
	started sync.Once
}

// NewVerifier creates and starts a new Verifier.
// If input or output channels are provided, they are created with VerifierChanSize capacity.
func NewVerifier(keys []crypto.PublicKey, input <-chan net.Message, output chan *consensus.Message) *Verifier {
	v := &Verifier{
		keys:   keys,
		input:  input,
		output: output,
	}
	v.cache, _ = simplelru.NewLRU(VerifierCacheSize, nil)
	if v.input == nil {
		v.input = make(<-chan net.Message, VerifierChanSize)
	}
	if v.output == nil {
		v.output = make(chan *consensus.Message, VerifierChanSize)
	}
	v.skipped = make(chan net.Message, VerifierChanSize)
	return v
}

// Input returns the channel of raw messages to be verified.
func (v *Verifier) Input() <-chan net.Message {
	return v.input
}

// Output returns the channel of verified consensus messages.
func (v *Verifier) Output() chan *consensus.Message {
	return v.output
}

// Skipped returns the channel of raw messages whose verification was skipped.
func (v *Verifier) Skipped() chan net.Message {
	return v.skipped
}

// Starts the verifier, namely its main routine in background.
// This main routine is only started in the first time this method is invoked.
func (v *Verifier) Start() {
	v.started.Do(func() { go v.mainRoutine() })
}

func (v *Verifier) mainRoutine() {
	for {
	NEXT_MESSAGE:
		select {
		case rawMessage := <-v.input:
			// Ignore any message that is not a consensus one
			if rawMessage.Code() != consensus.MessageCode {
				v.skipMessage(rawMessage)
				break NEXT_MESSAGE
			}
			message := consensus.MessageFromBytes(rawMessage)
			for _, sig := range message.GetCryptoSignatures() {
				key := sig.Key()
				v.stats.queries += 1
				// Skip the verification of signatures on the cache
				if _, exist := v.cache.Get(key); exist {
					v.stats.cached += 1
					continue
				}
				// If one signature verification fails, skip the message
				if !v.verifySignature(sig) {
					v.stats.rejected += 1
					break NEXT_MESSAGE
				}
				v.cache.Add(key, sig)
			}
			// If all signatures are valid, output the message
			v.output <- message
		}
	}
}

func (v *Verifier) verifySignature(sig *crypto.Signature) bool {
	if sig.ID < 0 || sig.ID >= len(v.keys) {
		return false
	}
	return v.keys[sig.ID].VerifySignature(sig.Payload, sig.Signature)
}

func (v *Verifier) skipMessage(message net.Message) {
	// Try to add the skipped message to the skipped channel
	select {
	case v.skipped <- message:
	default:
		// Do not block, discarding the message if the channel is full
		if message.Code() != bootstrap.MessageCode {
			log.Println("Verifier dropped message", message)
		}
	}
}

type VerifierStats struct {
	queries  int
	cached   int
	rejected int
}
