package babe

import (
	"bytes"
	"testing"

	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/stretchr/testify/assert"
)

var (
	nextConfigDescriptorBytes = []byte{0x1, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2}
)

func Test_NextConfigDescriptor_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := nextConfigDescriptor.Encode(buffer)
	assert.NoError(t, err)

	assert.Equal(t, nextConfigDescriptorBytes, buffer.Bytes())
}

func Test_NextConfigDescriptor_Bytes(t *testing.T) {
	assert.Equal(t, nextConfigDescriptorBytes, nextConfigDescriptor.Bytes())
}

func Test_DecodeNextConfigDescriptor(t *testing.T) {
	buffer := bytes.NewBuffer(nextConfigDescriptorBytes)

	result, err := DecodeNextConfigDescriptor(buffer)
	assert.NoError(t, err)

	assert.Equal(t, nextConfigDescriptor, result)
}

func Test_DecodeNextConfigDescriptor_Invalid(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x2})
	buffer.Write(nextConfigDescriptorBytes[1:])

	_, err := DecodeNextConfigDescriptor(buffer)

	assert.Equal(t, errInvalidNextConfigDescriptor, err)
}

func Test_DecodeNextConfigDescriptor_InvalidConfig(t *testing.T) {
	buffer := bytes.NewBuffer(nextConfigDescriptorBytes[:len(nextConfigDescriptorBytes)-1])
	buffer.Write([]byte{0x3})

	_, err := DecodeNextConfigDescriptor(buffer)

	assert.Equal(t, babetypes.ErrInvalidAllowedSlots, err)
}
