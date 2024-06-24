package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	expectedHrmpChannelUpdate, _ = hex.DecodeString("0500000006000000")
)

var (
	targetHrmpChannelUpdate = HrmpChannelUpdate{
		MsgCount:   5,
		TotalBytes: 6,
	}
)

func Test_HrmpChannelUpdate_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetHrmpChannelUpdate.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedHrmpChannelUpdate, buffer.Bytes())
}

func Test_HrmpChannelUpdate_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedHrmpChannelUpdate)

	result, err := DecodeHrmpChannelUpdate(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetHrmpChannelUpdate, result)
}

func Test_HrmpChannelUpdate_Bytes(t *testing.T) {
	assert.Equal(t, expectedHrmpChannelUpdate, targetHrmpChannelUpdate.Bytes())
}

func Test_HrmpChannelUpdate_IsEmpty(t *testing.T) {
	assert.False(t, targetHrmpChannelUpdate.IsEmpty())
}

func Test_HrmpChannelUpdate_Subtract(t *testing.T) {
	target := HrmpChannelUpdate{
		MsgCount:   2,
		TotalBytes: 3,
	}
	other := HrmpChannelUpdate{
		MsgCount:   1,
		TotalBytes: 2,
	}
	expect := HrmpChannelUpdate{
		MsgCount:   1,
		TotalBytes: 1,
	}

	target.Subtract(other)

	assert.Equal(t, expect, target)

}
