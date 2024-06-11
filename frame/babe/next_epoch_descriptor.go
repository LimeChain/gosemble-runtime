package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Information about the next epoch. This is broadcast in the first block
// of the epoch.
type NextEpochDescriptor struct {
	Authorities sc.Sequence[primitives.Authority]
	// The value of randomness to use for the slot-assignment.
	Randomness babetypes.Randomness
}

func (d NextEpochDescriptor) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		d.Authorities,
		d.Randomness,
	)
}

func (d NextEpochDescriptor) Bytes() []byte {
	return sc.EncodedBytes(d)
}
