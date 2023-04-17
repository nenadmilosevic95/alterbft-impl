package consensus

type Consensus interface {
	GetEpoch() int64

	Start(validCertificate *Certificate, lockedCertificate *Certificate)

	Stop()

	Started() bool

	ProcessMessage(message *Message)

	ProcessTimeout(timeout *Timeout)
}
