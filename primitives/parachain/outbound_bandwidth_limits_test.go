package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesOutboundBandwidthLimits, _ = hex.DecodeString("040000000400000004010000000300000004000000")
)

var (
	targetOutboundBandwidthLimits = OutboundBandwidthLimits{
		UmpMessagesRemaining: 3,
		UmpBytesRemaining:    4,
		HrmpOutgoing: sc.Dictionary[sc.U32, HrmpOutboundLimits]{
			1: HrmpOutboundLimits{
				BytesRemaining:    3,
				MessagesRemaining: 4,
			},
		},
	}
)

func Test_OutboundBandwidthLimits_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetOutboundBandwidthLimits.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesOutboundBandwidthLimits, buffer.Bytes())
}

func Test_OutboundBandwidthLimits_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesOutboundBandwidthLimits, targetOutboundBandwidthLimits.Bytes())
}
