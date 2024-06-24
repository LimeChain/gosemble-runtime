package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesHorizontalMessages, _ = hex.DecodeString("040500000004040000000c010203")
)

var (
	targetHorizontalMessages = HorizontalMessages{
		messages: sc.Dictionary[sc.U32, sc.Sequence[InboundDownwardMessage]]{
			5: sc.Sequence[InboundDownwardMessage]{
				{
					SentAt: 4,
					Msg:    sc.Sequence[sc.U8]{1, 2, 3},
				},
			},
		},
	}
)

func Test_HorizontalMessages_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetHorizontalMessages.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesHorizontalMessages, buffer.Bytes())
}

func Test_HorizontalMessages_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesHorizontalMessages)

	result, err := DecodeHorizontalMessages(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetHorizontalMessages, result)
}

func Test_HorizontalMessages_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesHorizontalMessages, targetHorizontalMessages.Bytes())
}

func Test_HorizontalMessages_UnprocessedMessages(t *testing.T) {
	expect := targetHorizontalMessages

	result := targetHorizontalMessages.UnprocessedMessages(1)

	assert.Equal(t, expect, result)
}

func Test_HorizontalMessages_UnprocessedMessages_Empty(t *testing.T) {
	expect := HorizontalMessages{sc.Dictionary[sc.U32, sc.Sequence[InboundDownwardMessage]]{
		5: sc.Sequence[InboundDownwardMessage]{},
	}}

	result := targetHorizontalMessages.UnprocessedMessages(10)

	assert.Equal(t, expect, result)
}
