package babe

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

func Test_NewPrimarySlots(t *testing.T) {
	assert.Equal(t, AllowedSlots{sc.NewVaryingData(sc.U8(0))}, NewPrimarySlots())
}

func Test_NewPrimaryAndSecondaryPlainSlots(t *testing.T) {
	assert.Equal(t, AllowedSlots{sc.NewVaryingData(sc.U8(1))}, NewPrimaryAndSecondaryPlainSlots())
}

func Test_NewPrimaryAndSecondaryVRFSlots(t *testing.T) {
	assert.Equal(t, AllowedSlots{sc.NewVaryingData(sc.U8(2))}, NewPrimaryAndSecondaryVRFSlots())
}

func Test_DecodeAllowedSlots(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       []byte
		expectation AllowedSlots
	}{
		{
			label:       "[]byte{0}",
			input:       []byte{0},
			expectation: NewPrimarySlots(),
		},
		{
			label:       "[]byte{1}",
			input:       []byte{1},
			expectation: NewPrimaryAndSecondaryPlainSlots(),
		},
		{
			label:       "[]byte{2}",
			input:       []byte{2},
			expectation: NewPrimaryAndSecondaryVRFSlots(),
		},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			buffer := bytes.NewBuffer(testExample.input)

			result, err := DecodeAllowedSlots(buffer)

			assert.NoError(t, err)
			assert.Equal(t, testExample.expectation, result)
		})
	}
}

func Test_DecodeAllowedSlots_Invalid(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{3})

	result, err := DecodeAllowedSlots(buffer)

	assert.Equal(t, err, ErrInvalidAllowedSlots)
	assert.Equal(t, AllowedSlots{}, result)
}

func Test_AllowedSlots_String(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       AllowedSlots
		expectation string
	}{
		{
			label:       "PrimarySlots",
			input:       NewPrimarySlots(),
			expectation: "PrimarySlots",
		},
		{
			label:       "PrimaryAndSecondaryPlainSlots",
			input:       NewPrimaryAndSecondaryPlainSlots(),
			expectation: "PrimaryAndSecondaryPlainSlots",
		},
		{
			label:       "PrimaryAndSecondaryVRFSlots",
			input:       NewPrimaryAndSecondaryVRFSlots(),
			expectation: "PrimaryAndSecondaryVRFSlots",
		},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			assert.Equal(t, testExample.expectation, testExample.input.String())
		})
	}
}

func Test_AllowedSlots_String_Invalid(t *testing.T) {
	result := AllowedSlots{sc.NewVaryingData(sc.U8(3))}.String()
	assert.Equal(t, "invalid representation of AllowedSlots", result)
}
