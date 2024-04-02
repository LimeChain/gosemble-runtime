package types

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// One of a number of consequences of depositing a fungible from an account.
type DepositConsequence sc.U8

const (
	// Deposit couldn't happen due to the amount being too low. This is usually because the
	// account doesn't yet exist and the deposit wouldn't bring it to at least the minimum needed
	// for existance.
	DepositConsequenceBelowMinimum DepositConsequence = iota
	// Deposit cannot happen since the account cannot be created (usually because it's a consumer
	// and there exists no provider reference).
	DepositConsequenceCannotCreate
	// The asset is unknown. Usually because an `AssetId` has been presented which doesn't exist
	// on the system.
	DepositConsequenceUnknownAsset
	// An overflow would occur. This is practically unexpected, but could happen in test systems
	// with extremely small balance types or balances that approach the max value of the balance
	// type.
	DepositConsequenceOverflow
	// Account continued in existence.
	DepositConsequenceSuccess
	// Account cannot receive the assets.
	DepositConsequenceBlocked
)

func (d DepositConsequence) ToResult() error {
	switch d {
	case DepositConsequenceBelowMinimum:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorNoFunds())
	case DepositConsequenceCannotCreate:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorWouldDie())
	case DepositConsequenceUnknownAsset:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorUnknownAsset())
	case DepositConsequenceOverflow:
		return primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
	case DepositConsequenceBlocked:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorBlocked())
	case DepositConsequenceSuccess:
		return nil
	default:
		return errors.New("Unknown DepositConsequence")
	}
}

func DecodeDepositConsequences(buffer *bytes.Buffer) (DepositConsequence, error) {
	value, err := sc.DecodeU8(buffer)
	if err != nil {
		return 0, err
	}

	switch d := DepositConsequence(value); d {
	case DepositConsequenceBelowMinimum, DepositConsequenceCannotCreate, DepositConsequenceUnknownAsset, DepositConsequenceOverflow, DepositConsequenceSuccess, DepositConsequenceBlocked:
		return d, nil
	default:
		return 0, errors.New("Invalid DepositConsequence")
	}
}
