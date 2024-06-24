package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesOutboundHrmpMessage, _ = hex.DecodeString("050000000c050607")
)

var (
	targetOutboundHrmpMessage = OutboundHrmpMessage{
		Id:   5,
		Data: sc.Sequence[sc.U8]{5, 6, 7},
	}
)

func Test_OutboundHrmpMessage_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetOutboundHrmpMessage.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesOutboundHrmpMessage, buffer.Bytes())
}

func Test_OutboundHrmpMessage_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesOutboundHrmpMessage)

	result, err := DecodeOutboundHrmpMessage(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetOutboundHrmpMessage, result)
}

func Test_OutboundHrmpMessage_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesOutboundHrmpMessage, targetOutboundHrmpMessage.Bytes())
}
