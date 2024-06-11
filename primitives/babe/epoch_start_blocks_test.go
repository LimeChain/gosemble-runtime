package babe

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	epochStartBlocks = EpochStartBlocks{
		Previous: 1,
		Current:  2,
	}
)

var (
	epochStartBlocksBytes = []byte{
		0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	}
)

func Test_EpochStartBlocks_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := epochStartBlocks.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, epochStartBlocksBytes, buffer.Bytes())
}

func Test_EpochStartBlocks_Bytes(t *testing.T) {
	assert.Equal(t, epochStartBlocksBytes, epochStartBlocks.Bytes())
}

func Test_DecodeEpochStartBlocks(t *testing.T) {
	buffer := bytes.NewBuffer(epochStartBlocksBytes)

	result, err := DecodeEpochStartBlocks(buffer)

	assert.NoError(t, err)
	assert.Equal(t, epochStartBlocks, result)
}
