package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Configuration data used by the BABE consensus engine that may change with epochs.
type EpochConfiguration struct {
	// A constant value that is used in the threshold calculation formula.
	// In the threshold formula calculation, `1 - c` represents the probability
	// of a slot being empty.
	C primitives.Tuple2U64

	// Whether this chain should run with secondary slots, which are assigned
	// in round-robin manner.
	AllowedSlots AllowedSlots
}

func (c EpochConfiguration) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		c.C,
		c.AllowedSlots,
	)
}

func (c EpochConfiguration) Bytes() []byte {
	return sc.EncodedBytes(c)
}

func DecodeEpochConfiguration(buffer *bytes.Buffer) (EpochConfiguration, error) {
	C, err := primitives.DecodeTuple2U64(buffer)
	if err != nil {
		return EpochConfiguration{}, err
	}

	AllowedSlots, err := DecodeAllowedSlots(buffer)
	if err != nil {
		return EpochConfiguration{}, err
	}

	return EpochConfiguration{
		C:            C,
		AllowedSlots: AllowedSlots,
	}, nil
}
