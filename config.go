package tendermint

import (
	"time"

	"dslab.inf.usi.ch/tendermint/crypto"
	"dslab.inf.usi.ch/tendermint/net"
)

// Config defines the configuration for Tendermint.
type Config struct {
	Log net.Log

	// BootstrapTickInterval is the clock tick period for the bootstrap
	// protocol. A process announces itself again after every interval.
	BootstrapTickInterval time.Duration

	// When set to a positive value, defines the maximum number of epochs to start.
	MaxEpochToStart int64

	// Maximum number of epochs started but not yet decided.
	MaxActiveEpochs int64

	// Maximum number of the blockchain heights we can have in the blockchain window.
	BlockchainSize int64

	// Model in which we operate.
	Model string

	// Byzantine processes.
	Byzantines map[int]bool

	// Time a byzantine leader should wait before proposing.
	ByzTime int

	// Byz attack
	ByzAttack string

	// Number of chunks in delta_chunked
	ChunksNumber int

	// Lower bound of the consensus instances window.
	// Prefix of decided instances of consensus to track.
	PastInstancesTracked int64

	// Upper bound of the consensus instances window.
	// Suffix of non-started instances of consensus to track.
	FutureInstancesTracked int64

	// Size (lenght) of internal message processing queues.
	MessageQueuesSize int

	// Schedule and trigger timeouts to the consensus protocol.
	// If set to false, the protocol will not tolerate failures.
	ScheduleTimeouts bool

	// If set to true, received messages have their signatures verified.
	// In this case, 'PublicKeys' should contain keys for every process.
	// Received messages with wrong or invalid signatures are discarded.
	VerifySignatures bool

	// Number of threads used for generating message signatures.
	// If unset, signatures are generated in the process main thread.
	SignatureGenerationThreads int

	// Number of threads used for verifying message signatures.
	// If unset, signatures are verified in the process main thread.
	SignatureVerificationThreads int

	// Used to produce signatures attached to generated messages.
	PrivateKeys []crypto.PrivateKey

	// Set of public keys used to verfiy signatures attached to messages.
	// Each process in the system should have an associated public key.
	PublicKeys []crypto.PublicKey

	// Defining maximum communication delay in synchronous system.
	TimeoutSmallDelta time.Duration
	TimeoutBigDelta   time.Duration
	FastAlterEnabled  bool

	// If set, defines the interval for publishing process stats.
	StatsPublishingInterval time.Duration
}

// DefaultConfig returns a default configuration for Tendermint.
func DefaultConfig() *Config {
	return &Config{
		BootstrapTickInterval: 100 * time.Millisecond,

		PastInstancesTracked:   16,
		FutureInstancesTracked: 16,
		MaxActiveEpochs:        2000,
		BlockchainSize:         2000,
		Model:                  "sync",

		Byzantines: nil,
		ByzTime:    0,

		MessageQueuesSize: 32,

		SignatureGenerationThreads:   0,
		SignatureVerificationThreads: 0,

		ScheduleTimeouts:  true,
		TimeoutSmallDelta: time.Second / 5,
		TimeoutBigDelta:   time.Second,

		FastAlterEnabled: false,

		ChunksNumber: 64,
	}
}
