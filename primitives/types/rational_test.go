package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	rationalValue = RationalValue{
		Numerator:   1,
		Denominator: 2,
	}
)

var (
	rationalValueBytes = []byte{
		0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	}
)

func Test_RationalValue_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := rationalValue.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, rationalValueBytes, buffer.Bytes())
}

func Test_RationalValue_Bytes(t *testing.T) {
	assert.Equal(t, rationalValueBytes, rationalValue.Bytes())
}

func Test_DecodeRationalValue(t *testing.T) {
	buffer := bytes.NewBuffer(rationalValueBytes)

	decodedRationalValue, err := DecodeRationalValue(buffer)

	assert.NoError(t, err)
	assert.Equal(t, rationalValue, decodedRationalValue)
}
