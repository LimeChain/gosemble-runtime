package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncodeDispatchError(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       DispatchError
		expectation []byte
	}{
		{label: "Encode(DispatchError('unknown error'))", input: NewDispatchErrorOther("unknown error"), expectation: []byte{0x00, 0x34, 0x75, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72}},
		{label: "Encode(DispatchErrorCannotLookup)", input: NewDispatchErrorCannotLookup(), expectation: []byte{0x01}},
		{label: "Encode(DispatchErrorBadOrigin)", input: NewDispatchErrorBadOrigin(), expectation: []byte{0x02}},
		{label: "Encode(DispatchErrorModule)", input: NewDispatchErrorModule(CustomModuleError{}), expectation: []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{label: "Encode(DispatchErrorConsumerRemaining)", input: NewDispatchErrorConsumerRemaining(), expectation: []byte{0x04}},
		{label: "Encode(DispatchErrorNoProviders)", input: NewDispatchErrorNoProviders(), expectation: []byte{0x05}},
		{label: "Encode(DispatchErrorTooManyConsumers)", input: NewDispatchErrorTooManyConsumers(), expectation: []byte{0x06}},
		{label: "Encode(DispatchErrorToken)", input: NewDispatchErrorToken(NewTokenErrorFundsUnavailable()), expectation: []byte{0x07, 0x00}},
		{label: "Encode(DispatchErrorArithmetic)", input: NewDispatchErrorArithmetic(NewArithmeticErrorUnderflow()), expectation: []byte{0x08, 0x00}},
		{label: "Encode(DispatchErrorTransactional)", input: NewDispatchErrorTransactional(NewTransactionalErrorLimitReached()), expectation: []byte{0x09, 0x00}},
		{label: "Encode(DispatchErrorExhausted)", input: NewDispatchErrorExhausted(), expectation: []byte{0xa}},
		{label: "Encode(DispatchErrorCorruption)", input: NewDispatchErrorCorruption(), expectation: []byte{0xb}},
		{label: "Encode(DispatchErrorUnavailable)", input: NewDispatchErrorUnavailable(), expectation: []byte{0xc}},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			buffer := &bytes.Buffer{}

			err := testExample.input.Encode(buffer)

			assert.NoError(t, err)
			assert.Equal(t, testExample.expectation, buffer.Bytes())
		})
	}
}

func Test_DecodeDispatchError(t *testing.T) {
	var testExamples = []struct {
		label       string
		input       []byte
		expectation DispatchError
	}{

		{label: "DecodeDispatchError(0x00, 0x34, 0x75, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72)", input: []byte{0x00, 0x34, 0x75, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72}, expectation: NewDispatchErrorOther("unknown error")},
		{label: "DecodeDispatchError(0x01)", input: []byte{0x01}, expectation: NewDispatchErrorCannotLookup()},
		{label: "DecodeDispatchError(0x02)", input: []byte{0x02}, expectation: NewDispatchErrorBadOrigin()},
		{label: "DecodeDispatchError(0x03, 0x00, 0x00, 0x00)", input: []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x00}, expectation: NewDispatchErrorModule(CustomModuleError{})},
		{label: "DecodeDispatchError(DispatchErrorConsumerRemaining)", input: []byte{0x04}, expectation: NewDispatchErrorConsumerRemaining()},
		{label: "DecodeDispatchError(DispatchErrorNoProviders)", input: []byte{0x05}, expectation: NewDispatchErrorNoProviders()},
		{label: "DecodeDispatchError(DispatchErrorTooManyConsumers)", input: []byte{0x06}, expectation: NewDispatchErrorTooManyConsumers()},
		{label: "DecodeDispatchError(DispatchErrorToken)", input: []byte{0x07, 0x00}, expectation: NewDispatchErrorToken(NewTokenErrorFundsUnavailable())},
		{label: "DecodeDispatchError(DispatchErrorArithmetic)", input: []byte{0x08, 0x00}, expectation: NewDispatchErrorArithmetic(NewArithmeticErrorUnderflow())},
		{label: "DecodeDispatchError(DispatchErrorTransactional)", input: []byte{0x09, 0x00}, expectation: NewDispatchErrorTransactional(NewTransactionalErrorLimitReached())},
		{label: "DecodeDispatchError(DispatchErrorExhausted)", input: []byte{0xa}, expectation: NewDispatchErrorExhausted()},
		{label: "DecodeDispatchError(DispatchErrorCorruption)", input: []byte{0xb}, expectation: NewDispatchErrorCorruption()},
		{label: "DecodeDispatchError(DispatchErrorUnavailable)", input: []byte{0xc}, expectation: NewDispatchErrorUnavailable()},
	}

	for _, testExample := range testExamples {
		t.Run(testExample.label, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			buffer.Write(testExample.input)

			result, err := DecodeDispatchError(buffer)
			assert.NoError(t, err)

			assert.Equal(t, testExample.expectation, result)
			assert.NotEmpty(t, result.Error())
		})
	}
}
