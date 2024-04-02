package types

import (
	"bytes"
	"errors"
	"testing"

	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func TestDepositConsequence_ToResult(t *testing.T) {
	tests := []struct {
		name               string
		depositConsequence DepositConsequence
		expectedErr        error
	}{
		{
			name:               "DepositConsequenceBelowMinimum",
			depositConsequence: DepositConsequenceBelowMinimum,
			expectedErr:        primitives.NewDispatchErrorToken(primitives.NewTokenErrorNoFunds()),
		},
		{
			name:               "DepositConsequenceCannotCreate",
			depositConsequence: DepositConsequenceCannotCreate,
			expectedErr:        primitives.NewDispatchErrorToken(primitives.NewTokenErrorWouldDie()),
		},
		{
			name:               "DepositConsequenceUnknownAsset",
			depositConsequence: DepositConsequenceUnknownAsset,
			expectedErr:        primitives.NewDispatchErrorToken(primitives.NewTokenErrorUnknownAsset()),
		},
		{
			name:               "DepositConsequenceOverflow",
			depositConsequence: DepositConsequenceOverflow,
			expectedErr:        primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow()),
		},
		{
			name:               "DepositConsequenceSuccess",
			depositConsequence: DepositConsequenceSuccess,
			expectedErr:        nil,
		},
		{
			name:               "DepositConsequenceBlocked",
			depositConsequence: DepositConsequenceBlocked,
			expectedErr:        primitives.NewDispatchErrorToken(primitives.NewTokenErrorBlocked()),
		},
		{
			name:               "UnknownDepositConsequence",
			depositConsequence: DepositConsequence(100),
			expectedErr:        errors.New("Unknown DepositConsequence"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.depositConsequence.ToResult()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestDecodeDepositConsequences(t *testing.T) {
	tests := []struct {
		name           string
		buffer         *bytes.Buffer
		expectedResult DepositConsequence
		expectedError  error
	}{
		{
			name:           "ValidDepositConsequenceBelowMinimum",
			buffer:         bytes.NewBuffer([]byte{0}),
			expectedResult: DepositConsequenceBelowMinimum,
			expectedError:  nil,
		},
		{
			name:           "ValidDepositConsequenceCannotCreate",
			buffer:         bytes.NewBuffer([]byte{1}),
			expectedResult: DepositConsequenceCannotCreate,
			expectedError:  nil,
		},
		{
			name:           "ValidDepositConsequenceUnknownAsset",
			buffer:         bytes.NewBuffer([]byte{2}),
			expectedResult: DepositConsequenceUnknownAsset,
			expectedError:  nil,
		},
		{
			name:           "ValidDepositConsequenceOverflow",
			buffer:         bytes.NewBuffer([]byte{3}),
			expectedResult: DepositConsequenceOverflow,
			expectedError:  nil,
		},
		{
			name:           "ValidDepositConsequenceSuccess",
			buffer:         bytes.NewBuffer([]byte{4}),
			expectedResult: DepositConsequenceSuccess,
			expectedError:  nil,
		},
		{
			name:           "ValidDepositConsequenceBlocked",
			buffer:         bytes.NewBuffer([]byte{5}),
			expectedResult: DepositConsequenceBlocked,
			expectedError:  nil,
		},
		{
			name:           "InvalidDepositConsequence",
			buffer:         bytes.NewBuffer([]byte{100}),
			expectedResult: 0,
			expectedError:  errors.New("Invalid DepositConsequence"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeDepositConsequences(tt.buffer)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
