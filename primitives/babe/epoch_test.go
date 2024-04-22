package babe

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	epoch = Epoch{
		EpochIndex:  sc.U64(1),
		StartSlot:   Slot(2),
		Duration:    sc.U64(3),
		Authorities: sc.Sequence[primitives.Authority]{},
		Randomness:  NewRandomness(),
		Config: EpochConfiguration{
			C:            primitives.RationalValue{Numerator: 2, Denominator: 3},
			AllowedSlots: NewPrimarySlots(),
		},
	}
)

var (
	epochBytes = []byte{
		0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // EpochIndex
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // StartSlot
		0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // Duration
		0x0,                                                                                                                                                            // Authorities
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // Randomness
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // C1
		0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // C2
		0x0, // AllowedSlots
	}
)

func Test_Epoch_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	epoch.Encode(buffer)

	assert.Equal(t, epochBytes, epoch.Bytes())
}

func Test_Epoch_Bytes(t *testing.T) {
	assert.Equal(t, epochBytes, epoch.Bytes())
}

func Test_DecodeEpoch(t *testing.T) {
	buffer := bytes.NewBuffer(epochBytes)

	result, err := DecodeEpoch(buffer)

	assert.Nil(t, err)
	assert.Equal(t, epoch, result)
}
