package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Configuration data used by the BABE consensus engine.
type Configuration struct {
	// The slot duration in milliseconds for BABE. Currently, only
	// the value provided by this type at genesis will be used.
	//
	// Dynamic slot duration may be supported in the future.
	SlotDuration sc.U64

	// The duration of epochs in slots.
	EpochLength sc.U64

	// A constant value that is used in the threshold calculation formula.
	// In the threshold formula calculation, `1 - c` represents the probability
	// of a slot being empty.
	C primitives.Tuple2U64

	// The authorities
	Authorities sc.Sequence[primitives.Authority]

	// The randomness
	Randomness Randomness

	// Type of allowed slots.
	AllowedSlots AllowedSlots
}

func (c Configuration) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		c.SlotDuration,
		c.EpochLength,
		c.C,
		c.Authorities,
		c.Randomness,
		c.AllowedSlots,
	)
}

func (c Configuration) Bytes() []byte {
	return sc.EncodedBytes(c)
}

func DecodeConfiguration(buffer *bytes.Buffer) (Configuration, error) {
	slotDuration, err := sc.DecodeU64(buffer)
	if err != nil {
		return Configuration{}, err
	}

	epochLength, err := sc.DecodeU64(buffer)
	if err != nil {
		return Configuration{}, err
	}

	C, err := primitives.DecodeTuple2U64(buffer)
	if err != nil {
		return Configuration{}, err
	}

	authorities, err := sc.DecodeSequenceWith(buffer, primitives.DecodeAuthority)
	if err != nil {
		return Configuration{}, err
	}

	randomness, err := sc.DecodeFixedSequence[sc.U8](RandomnessLength, buffer)
	if err != nil {
		return Configuration{}, err
	}

	allowedSlots, err := DecodeAllowedSlots(buffer)
	if err != nil {
		return Configuration{}, err
	}

	return Configuration{
		SlotDuration: slotDuration,
		EpochLength:  epochLength,
		C:            C,
		Authorities:  authorities,
		Randomness:   randomness,
		AllowedSlots: allowedSlots,
	}, nil
}
