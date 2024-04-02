package types

import (
	"bytes"
	"errors"
	"testing"

	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawConsequence_ToResult(t *testing.T) {
	tests := []struct {
		name         string
		conseq       WithdrawConsequence
		preservation Preservation
		expected     error
	}{
		{
			name:     "WithdrawConsequenceBalanceLow",
			conseq:   WithdrawConsequenceBalanceLow,
			expected: primitives.NewDispatchErrorToken(primitives.NewTokenErrorNoFunds()),
		},
		{
			name:     "WithdrawConsequenceWouldDie",
			conseq:   WithdrawConsequenceWouldDie,
			expected: primitives.NewDispatchErrorToken(primitives.NewTokenErrorWouldDie()),
		},
		{
			name:     "WithdrawConsequenceUnknownAsset",
			conseq:   WithdrawConsequenceUnknownAsset,
			expected: primitives.NewDispatchErrorToken(primitives.NewTokenErrorUnknownAsset()),
		},
		{
			name:     "WithdrawConsequenceUnderflow",
			conseq:   WithdrawConsequenceUnderflow,
			expected: primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorUnderflow()),
		},
		{
			name:     "WithdrawConsequenceOverflow",
			conseq:   WithdrawConsequenceOverflow,
			expected: primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow()),
		},
		{
			name:     "WithdrawConsequenceFrozen",
			conseq:   WithdrawConsequenceFrozen,
			expected: primitives.NewDispatchErrorToken(primitives.NewTokenErrorFrozen()),
		},
		{
			name:         "WithdrawConsequenceReducedToZero Not Expendable",
			conseq:       WithdrawConsequenceReducedToZero,
			preservation: PreservationPreserve,
			expected:     primitives.NewDispatchErrorToken(primitives.NewTokenErrorNotExpendable()),
		},
		{
			name:     "WithdrawConsequenceReducedToZero",
			conseq:   WithdrawConsequenceReducedToZero,
			expected: nil,
		},
		{
			name:     "WithdrawConsequenceSuccess",
			conseq:   WithdrawConsequenceSuccess,
			expected: nil,
		},
		{
			name:         "Unknown WithdrawConsequence",
			conseq:       WithdrawConsequence(10),
			preservation: PreservationExpendable,
			expected:     errors.New("Unknown WithdrawConsequence"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conseq.ToResult(tt.preservation)
			assert.Equal(t, tt.expected, err)
		})
	}
}

func TestDecodeWithdrawConsequences(t *testing.T) {
	tests := []struct {
		name     string
		buffer   *bytes.Buffer
		expected WithdrawConsequence
		err      error
	}{
		{
			name:     "Valid WithdrawConsequence",
			buffer:   bytes.NewBuffer([]byte{0}),
			expected: WithdrawConsequenceBalanceLow,
			err:      nil,
		},
		{
			name:     "Invalid WithdrawConsequence",
			buffer:   bytes.NewBuffer([]byte{10}),
			expected: 0,
			err:      errors.New("Invalid WithdrawConsequence"),
		},
		{
			name:     "Cannot Decode sc.U8",
			buffer:   &bytes.Buffer{},
			expected: 0,
			err:      errors.New("EOF"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conseq, err := DecodeWithdrawConsequences(tt.buffer)
			assert.Equal(t, tt.expected, conseq)
			assert.Equal(t, tt.err, err)
		})
	}
}
