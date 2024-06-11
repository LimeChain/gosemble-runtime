package types

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type DepositConsequence struct {
	sc.VaryingData
}

const (
	DepositConsequenceBelowMinimum sc.U8 = iota
	DepositConsequenceCannotCreate
	DepositConsequenceUnknownAsset
	DepositConsequenceOverflow
	DepositConsequenceSuccess
	DepositConsequenceBlocked
)

func NewDepositConsequenceBelowMinimum() DepositConsequence {
	return DepositConsequence{sc.NewVaryingData(DepositConsequenceBelowMinimum)}
}

func NewDepositConsequenceCannotCreate() DepositConsequence {
	return DepositConsequence{sc.NewVaryingData(DepositConsequenceCannotCreate)}
}

func NewDepositConsequenceUnknownAsset() DepositConsequence {
	return DepositConsequence{sc.NewVaryingData(DepositConsequenceUnknownAsset)}
}

func NewDepositConsequenceOverflow() DepositConsequence {
	return DepositConsequence{sc.NewVaryingData(DepositConsequenceOverflow)}
}

func NewDepositConsequenceSuccess() DepositConsequence {
	return DepositConsequence{sc.NewVaryingData(DepositConsequenceSuccess)}
}

func NewDepositConsequenceBlocked() DepositConsequence {
	return DepositConsequence{sc.NewVaryingData(DepositConsequenceBlocked)}
}

func (wc DepositConsequence) IntoResult() error {
	switch wc.VaryingData[0] {
	case DepositConsequenceBelowMinimum:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorBelowMinimum())
	case DepositConsequenceCannotCreate:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorCannotCreate())
	case DepositConsequenceUnknownAsset:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorUnknownAsset())
	case DepositConsequenceOverflow:
		return primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
	case DepositConsequenceBlocked:
		return primitives.NewDispatchErrorToken(primitives.NewTokenErrorBlocked())
	case DepositConsequenceSuccess:
		return nil
	default:
		return primitives.NewDispatchErrorOther("invalid DepositConsequence type")
	}
}
