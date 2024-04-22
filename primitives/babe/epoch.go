package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// BABE epoch information
type Epoch struct {
	// The epoch index.
	EpochIndex sc.U64
	// The starting slot of the epoch.
	StartSlot Slot
	// The duration of this epoch.
	Duration sc.U64
	// The authorities and their weights.
	Authorities sc.Sequence[primitives.Authority]
	// Randomness for this epoch.
	Randomness Randomness
	// Configuration of the epoch.
	Config EpochConfiguration
}

func (e Epoch) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		e.EpochIndex,
		e.StartSlot,
		e.Duration,
		e.Authorities,
		e.Randomness,
		e.Config,
	)
}

func (e Epoch) Bytes() []byte {
	return sc.EncodedBytes(e)
}

func DecodeEpoch(buffer *bytes.Buffer) (Epoch, error) {
	epochIndex, err := sc.DecodeU64(buffer)
	if err != nil {
		return Epoch{}, err
	}

	startSlot, err := sc.DecodeU64(buffer)
	if err != nil {
		return Epoch{}, err
	}

	duration, err := sc.DecodeU64(buffer)
	if err != nil {
		return Epoch{}, err
	}

	authorities, err := sc.DecodeSequenceWith(buffer, primitives.DecodeAuthority)
	if err != nil {
		return Epoch{}, err
	}

	randomness, err := sc.DecodeFixedSequence[sc.U8](RandomnessLength, buffer)
	if err != nil {
		return Epoch{}, err
	}

	config, err := DecodeEpochConfiguration(buffer)
	if err != nil {
		return Epoch{}, err
	}

	return Epoch{
		EpochIndex:  epochIndex,
		StartSlot:   startSlot,
		Duration:    duration,
		Authorities: authorities,
		Randomness:  randomness,
		Config:      config,
	}, nil
}
