package types

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// One of a number of consequences of withdrawing a fungible from an account.
type WithdrawConsequence sc.U8

const (
	// Withdraw could not happen since the amount to be withdrawn is less than the total funds in
	// the account.
	WithdrawConsequenceBalanceLow WithdrawConsequence = iota
	// The withdraw would mean the account dying when it needs to exist (usually because it is a
	// provider and there are consumer references on it).
	WithdrawConsequenceWouldDie
	// The asset is unknown. Usually because an `AssetId` has been presented which doesn't exist
	// on the system.
	WithdrawConsequenceUnknownAsset
	// There has been an underflow in the system. This is indicative of a corrupt state and
	// likely unrecoverable.
	WithdrawConsequenceUnderflow
	// There has been an overflow in the system. This is indicative of a corrupt state and
	// likely unrecoverable.
	WithdrawConsequenceOverflow
	// Not enough of the funds in the account are unavailable for withdrawal.
	WithdrawConsequenceFrozen
	// Account balance would reduce to zero, potentially destroying it. The parameter is the
	// amount of balance which is destroyed.
	WithdrawConsequenceReducedToZero
	// Account continued in existence.
	WithdrawConsequenceSuccess
)

// Convert the type into a `Result` with `DispatchError` as the error or the additional
// `Balance` by which the account will be reduced.
func (w WithdrawConsequence) ToResult(preservation Preservation) error {
	switch w {
	case WithdrawConsequenceBalanceLow:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorNoFunds())
	case WithdrawConsequenceWouldDie:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorWouldDie())
	case WithdrawConsequenceUnknownAsset:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorUnknownAsset())
	case WithdrawConsequenceUnderflow:
		return primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorUnderflow())
	case WithdrawConsequenceOverflow:
		return primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
	case WithdrawConsequenceFrozen:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorFrozen())
	case WithdrawConsequenceReducedToZero:
		if preservation != PreservationExpendable {
			return primitives.NewDispatchErrorToken(primitives.NewTokenErrorNotExpendable())
		}
		return nil
	case WithdrawConsequenceSuccess:
		return nil
	default:
		return errors.New("Unknown WithdrawConsequence")
	}
}

func DecodeWithdrawConsequences(buffer *bytes.Buffer) (WithdrawConsequence, error) {
	value, err := sc.DecodeU8(buffer)
	if err != nil {
		return 0, err
	}

	switch w := WithdrawConsequence(value); w {
	case WithdrawConsequenceBalanceLow, WithdrawConsequenceWouldDie, WithdrawConsequenceUnknownAsset, WithdrawConsequenceUnderflow, WithdrawConsequenceOverflow, WithdrawConsequenceFrozen, WithdrawConsequenceReducedToZero, WithdrawConsequenceSuccess:
		return w, nil
	default:
		return 0, errors.New("Invalid WithdrawConsequence")
	}
}
