package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesInboundDownwardMessage, _ = hex.DecodeString("0a000000080506")
)

var (
	targetInboundDownwardMessage = InboundDownwardMessage{
		SentAt: 10,
		Msg:    sc.Sequence[sc.U8]{5, 6},
	}
)

func Test_InboundDownwardMessage_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetInboundDownwardMessage.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesInboundDownwardMessage, buffer.Bytes())
}

func Test_InboundDownwardMessage_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesInboundDownwardMessage)

	result, err := DecodeInboundDownwardMessage(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetInboundDownwardMessage, result)
}

func Test_InboundDownwardMessage_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesInboundDownwardMessage, targetInboundDownwardMessage.Bytes())
}
