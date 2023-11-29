package consensus

type Consensus interface {
	GetEpoch() int64

	Start(lockedCertificate *Certificate, sentLockedCertificate bool)

	Stop()

	Started() bool

	ProcessMessage(message *Message)

	ProcessTimeout(timeout *Timeout)
}
