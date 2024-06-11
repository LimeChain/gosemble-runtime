package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	unknownTransactionCannotLookup = NewTransactionValidityError(NewUnknownTransactionCannotLookup())
)

func Test_NewTransactionValidityError_TypeError(t *testing.T) {
	result := NewTransactionValidityError(sc.U8(6))

	assert.Equal(t, "not a valid 'TransactionValidityError' type", result.Error())
}

func Test_TransactionValidityError_Encode(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       error
		expectation []byte
	}{
		{
			label:       "Encode(TransactionValidityError(InvalidTransaction(PaymentError)))",
			input:       invalidTransactionPayment,
			expectation: []byte{0x00, 0x01},
		},
		{
			label:       "Encode(TransactionValidityError(UnknownTransaction(0)))",
			input:       unknownTransactionCannotLookup,
			expectation: []byte{0x01, 0x00},
		},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			buffer := &bytes.Buffer{}

			err := testExample.input.(TransactionValidityError).Encode(buffer)

			assert.NoError(t, err)
			assert.Equal(t, testExample.expectation, buffer.Bytes())
		})
	}
}

func Test_TransactionValidityError_Encode_TypeError(t *testing.T) {
	buffer := &bytes.Buffer{}

	tve := TransactionValidityError{sc.NewVaryingData(sc.U8(6))}

	err := tve.Encode(buffer)

	assert.Error(t, err)
	assert.Equal(t, "not a valid 'TransactionValidityError' type", err.Error())
	assert.Equal(t, &bytes.Buffer{}, buffer)
}

func Test_DecodeTransactionValidityError(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       []byte
		expectation error
	}{
		{
			label:       "DecodeTransactionValidityError(TransactionValidityError(InvalidTransaction(PaymentError)))",
			input:       []byte{0x00, 0x01},
			expectation: invalidTransactionPayment,
		},
		{
			label:       "DecodeTransactionValidityError(TransactionValidityError(UnknownTransaction(0)))",
			input:       []byte{0x01, 0x00},
			expectation: unknownTransactionCannotLookup,
		},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			buffer.Write(testExample.input)

			result, err := DecodeTransactionValidityError(buffer)
			assert.NoError(t, err)

			assert.Equal(t, testExample.expectation, result)
			assert.NotEmpty(t, result.Error())
		})
	}
}

func Test_DecodeTransactionValidityError_TypeError(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(6)

	res, err := DecodeTransactionValidityError(buffer)

	assert.Error(t, err)
	assert.Equal(t, "not a valid 'TransactionValidityError' type", err.Error())
	assert.Equal(t, TransactionValidityError{}, res)
}

func Test_TransactionValidityError_Bytes(t *testing.T) {
	expect, _ := hex.DecodeString("0001")

	assert.Equal(t, expect, invalidTransactionPayment.(TransactionValidityError).Bytes())
}
