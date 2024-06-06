package grandpa

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// A scheduled change of authority set.
type ScheduledChange struct {
	// The new authorities after the change, along with their respective weights.
	NextAuthorities sc.Sequence[primitives.Authority]
	// The number of blocks to delay.
	Delay sc.U64
}

func (s ScheduledChange) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		s.NextAuthorities,
		s.Delay,
	)
}

func (s ScheduledChange) Bytes() []byte {
	return sc.EncodedBytes(s)
}

func DecodeScheduledChange(buffer *bytes.Buffer) (ScheduledChange, error) {
	nextAuthorities, err := sc.DecodeSequenceWith(buffer, primitives.DecodeAuthority)
	if err != nil {
		return ScheduledChange{}, err
	}

	delay, err := sc.DecodeU64(buffer)
	if err != nil {
		return ScheduledChange{}, err
	}

	return ScheduledChange{
		NextAuthorities: nextAuthorities,
		Delay:           delay,
	}, nil
}
