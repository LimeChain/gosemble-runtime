package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesRelayDispatchQueueRemainingCapacity, _ = hex.DecodeString("0600000007000000")
)

var (
	targetRelayDispatchQueueRemainingCapacity = RelayDispatchQueueRemainingCapacity{
		RemainingCount: 6,
		RemainingSize:  7,
	}
)

func Test_RelayDispatchQueueRemainingCapacity_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetRelayDispatchQueueRemainingCapacity.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesRelayDispatchQueueRemainingCapacity, buffer.Bytes())
}

func Test_RelayDispatchQueueRemainingCapacity_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesRelayDispatchQueueRemainingCapacity)

	result, err := DecodeRelayDispatchQueueRemainingCapacity(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetRelayDispatchQueueRemainingCapacity, result)
}

func Test_RelayDispatchQueueRemainingCapacity_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesRelayDispatchQueueRemainingCapacity, targetRelayDispatchQueueRemainingCapacity.Bytes())
}
