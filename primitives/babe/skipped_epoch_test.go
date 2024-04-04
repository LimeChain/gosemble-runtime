package babe

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	skippedEpoch = SkippedEpoch{U64: 1, SessionIndex: 2}
)

var (
	skippedEpochBytes = []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0}
)

func Test_SkippedEpoch_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := skippedEpoch.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, skippedEpochBytes, buffer.Bytes())
}

func Test_SkippedEpoch_Bytes(t *testing.T) {
	assert.Equal(t, skippedEpochBytes, skippedEpoch.Bytes())
}
