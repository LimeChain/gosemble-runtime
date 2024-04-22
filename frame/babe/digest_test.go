package babe

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	primaryPreDigest = PrimaryPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
		VrfSignature:   vrfSignature,
	}

	secondaryPlainPreDigest = SecondaryPlainPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
	}

	secondaryVRFPreDigest = SecondaryVRFPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
		VrfSignature:   vrfSignature,
	}
)

var (
	expectedPrimaryPreDigest        = PreDigest{sc.NewVaryingData(Primary, primaryPreDigest)}
	expectedSecondaryPlainPreDigest = PreDigest{sc.NewVaryingData(SecondaryPlain, secondaryPlainPreDigest)}
	expectedSecondaryVRFPreDigest   = PreDigest{sc.NewVaryingData(SecondaryVRF, secondaryVRFPreDigest)}

	expectedPrimaryPreDigestBytes        = []byte{1, 1, 0, 0, 0, 130, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	expectedSecondaryPlainPreDigestBytes = []byte{2, 1, 0, 0, 0, 130, 0, 0, 0, 0, 0, 0, 0}
	expectedSecondaryVRFPreDigestBytes   = []byte{3, 1, 0, 0, 0, 130, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

func Test_Babe_NewPrimaryPreDigest(t *testing.T) {
	signature, err := types.NewVrfSignature(output, proof)
	assert.NoError(t, err)

	assert.Equal(t, expectedPrimaryPreDigest, NewPrimaryPreDigest(authorityIndex, slot, signature))
}

func Test_Babe_PrimaryPreDigest_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := expectedPrimaryPreDigest.Encode(buffer)
	assert.NoError(t, err)

	assert.Equal(t, expectedPrimaryPreDigestBytes, buffer.Bytes())
}

func Test_Babe_PrimaryPreDigest_Bytes(t *testing.T) {
	assert.Equal(t, expectedPrimaryPreDigestBytes, expectedPrimaryPreDigest.Bytes())
}

func Test_Babe_DecodePrimaryPreDigest(t *testing.T) {
	buffer := bytes.NewBuffer(expectedPrimaryPreDigest.Bytes())

	preDigest, err := DecodePreDigest(buffer)
	assert.NoError(t, err)

	assert.Equal(t, expectedPrimaryPreDigest, preDigest)
}

func Test_Babe_NewSecondaryPlainPreDigest(t *testing.T) {
	assert.Equal(t, expectedSecondaryPlainPreDigest, NewSecondaryPlainPreDigest(authorityIndex, slot))
}

func Test_Babe_SecondaryPlainPreDigest_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := expectedSecondaryPlainPreDigest.Encode(buffer)
	assert.NoError(t, err)

	assert.Equal(t, expectedSecondaryPlainPreDigestBytes, buffer.Bytes())
}

func Test_Babe_SecondaryPlainPreDigest_Bytes(t *testing.T) {
	assert.Equal(t, expectedSecondaryPlainPreDigestBytes, expectedSecondaryPlainPreDigest.Bytes())
}

func Test_Babe_DecodeSecondaryPlainPreDigest(t *testing.T) {
	buffer := bytes.NewBuffer(expectedSecondaryPlainPreDigest.Bytes())

	preDigest, err := DecodePreDigest(buffer)
	assert.NoError(t, err)

	assert.Equal(t, expectedSecondaryPlainPreDigest, preDigest)
}

func Test_Babe_NewSecondaryVRFPreDigest(t *testing.T) {
	signature, err := types.NewVrfSignature(output, proof)
	assert.NoError(t, err)

	assert.Equal(t, expectedSecondaryVRFPreDigest, NewSecondaryVRFPreDigest(authorityIndex, slot, signature))
}

func Test_Babe_SecondaryVRFPreDigest_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := expectedSecondaryVRFPreDigest.Encode(buffer)
	assert.NoError(t, err)

	assert.Equal(t, expectedSecondaryVRFPreDigestBytes, buffer.Bytes())
}

func Test_Babe_SecondaryVRFPreDigest_Bytes(t *testing.T) {
	assert.Equal(t, expectedSecondaryVRFPreDigestBytes, expectedSecondaryVRFPreDigest.Bytes())
}

func Test_Babe_DecodeSecondaryVRFPreDigest(t *testing.T) {
	buffer := bytes.NewBuffer(expectedSecondaryVRFPreDigest.Bytes())

	preDigest, err := DecodePreDigest(buffer)
	assert.NoError(t, err)

	assert.Equal(t, expectedSecondaryVRFPreDigest, preDigest)
}

func Test_Babe_AuthorityIndex(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       PreDigest
		expectation sc.U32
	}{
		{
			label:       "PrimaryPreDigest",
			input:       expectedPrimaryPreDigest,
			expectation: authorityIndex,
		},
		{
			label:       "SecondaryPlainPreDigest",
			input:       expectedSecondaryPlainPreDigest,
			expectation: authorityIndex,
		},
		{
			label:       "SecondaryVRFPreDigest",
			input:       expectedSecondaryVRFPreDigest,
			expectation: authorityIndex,
		},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			index, err := testExample.input.AuthorityIndex()
			assert.NoError(t, err)
			assert.Equal(t, testExample.expectation, index)
		})
	}
}

func Test_Babe_Slot(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       PreDigest
		expectation sc.U64
	}{
		{
			label:       "PrimaryPreDigest",
			input:       expectedPrimaryPreDigest,
			expectation: slot,
		},
		{
			label:       "SecondaryPlainPreDigest",
			input:       expectedSecondaryPlainPreDigest,
			expectation: slot,
		},
		{
			label:       "SecondaryVRFPreDigest",
			input:       expectedSecondaryVRFPreDigest,
			expectation: slot,
		},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			slotNumber, err := testExample.input.Slot()
			assert.NoError(t, err)
			assert.Equal(t, testExample.expectation, slotNumber)
		})
	}
}

func Test_Babe_VrfSignature(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       PreDigest
		expectation sc.Option[types.VrfSignature]
	}{
		{
			label:       "PrimaryPreDigest",
			input:       expectedPrimaryPreDigest,
			expectation: sc.NewOption[types.VrfSignature](vrfSignature),
		},
		{
			label:       "SecondaryPlainPreDigest",
			input:       expectedSecondaryPlainPreDigest,
			expectation: sc.NewOption[types.VrfSignature](nil),
		},
		{
			label:       "SecondaryVRFPreDigest",
			input:       expectedSecondaryVRFPreDigest,
			expectation: sc.NewOption[types.VrfSignature](vrfSignature),
		},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			signature, err := testExample.input.VrfSignature()
			assert.NoError(t, err)
			assert.Equal(t, testExample.expectation, signature)
		})
	}
}
