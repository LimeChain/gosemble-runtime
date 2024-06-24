package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesHrmpOutboundLimits, _ = hex.DecodeString("0100000002000000")
)

var (
	targetHrmpOutboundLimits = HrmpOutboundLimits{
		BytesRemaining:    1,
		MessagesRemaining: 2,
	}
)

func Test_HrmpOutboundLimits_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetHrmpOutboundLimits.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesHrmpOutboundLimits, buffer.Bytes())
}

func Test_HrmpOutboundLimits_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesHrmpOutboundLimits)

	result, err := DecodeHrmpOutboundLimits(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetHrmpOutboundLimits, result)
}

func Test_HrmpOutboundLimits_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesHrmpOutboundLimits, targetHrmpOutboundLimits.Bytes())
}
