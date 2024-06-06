package babe

import (
	"bytes"
	"testing"

	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	epochConfiguration = EpochConfiguration{
		C:            primitives.Tuple2U64{First: 2, Second: 3},
		AllowedSlots: NewPrimarySlots(),
	}
)

var (
	epochConfigurationBytes = []byte{
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0,
	}
)

func Test_EpochConfiguration_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := epochConfiguration.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, epochConfigurationBytes, buffer.Bytes())
}

func Test_EpochConfiguration_Bytes(t *testing.T) {
	assert.Equal(t, epochConfigurationBytes, epochConfiguration.Bytes())
}

func Test_DecodeEpochConfiguration(t *testing.T) {
	buffer := bytes.NewBuffer(epochConfigurationBytes)

	result, err := DecodeEpochConfiguration(buffer)

	assert.NoError(t, err)
	assert.Equal(t, epochConfiguration, result)
}
