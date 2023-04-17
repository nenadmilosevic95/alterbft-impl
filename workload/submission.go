package workload

import (
	"encoding/binary"
	"time"
)

// Encoding to store numeric fields in byte arrays
var encoding binary.ByteOrder = binary.LittleEndian

// Submission records the submission of a value.
type Submission struct {
	Sender int
	Size   int
	Time   time.Time
}

// NewSubmission creates a submission, setting its Time to now.
func NewSubmission(sender, size int) *Submission {
	return &Submission{
		Sender: sender,
		Size:   size,
		Time:   time.Now(),
	}
}

// SubmissionFromValue reads submission data from a value.
func SubmissionFromValue(value []byte) *Submission {
	s := new(Submission)
	s.Sender = int(encoding.Uint32(value[0:4]))
	s.Size = len(value)
	timestamp := &s.Time
	timestamp.UnmarshalBinary(value[4:19])
	return s
}

// Write the submission data to a submitted value.
func (s *Submission) Write(value []byte) {
	timestampBytes, _ := s.Time.MarshalBinary()
	encoding.PutUint32(value[0:4], uint32(s.Sender))
	copy(value[4:19], timestampBytes)
}
