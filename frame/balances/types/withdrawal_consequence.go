package types

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type WithdrawalConsequence struct {
	sc.VaryingData
}

const (
	WithdrawalConsequenceBalanceLow sc.U8 = iota
	WithdrawalConsequenceWouldDie
	WithdrawalConsequenceUnknownAsset
	WithdrawalConsequenceUnderflow
	WithdrawalConsequenceOverflow
	WithdrawalConsequenceFrozen
	WithdrawalConsequenceReducedToZero
	WithdrawalConsequenceSuccess
)

func NewWithdrawalConsequenceBalanceLow() WithdrawalConsequence {
	return WithdrawalConsequence{sc.NewVaryingData(WithdrawalConsequenceBalanceLow)}
}

func NewWithdrawalConsequenceWouldDie() WithdrawalConsequence {
	return WithdrawalConsequence{sc.NewVaryingData(WithdrawalConsequenceWouldDie)}
}

func NewWithdrawalConsequenceUnknownAsset() WithdrawalConsequence {
	return WithdrawalConsequence{sc.NewVaryingData(WithdrawalConsequenceUnknownAsset)}
}

func NewWithdrawalConsequenceUnderflow() WithdrawalConsequence {
	return WithdrawalConsequence{sc.NewVaryingData(WithdrawalConsequenceUnderflow)}
}

func NewWithdrawalConsequenceOverflow() WithdrawalConsequence {
	return WithdrawalConsequence{sc.NewVaryingData(WithdrawalConsequenceOverflow)}
}

func NewWithdrawalConsequenceFrozen() WithdrawalConsequence {
	return WithdrawalConsequence{sc.NewVaryingData(WithdrawalConsequenceFrozen)}
}

func NewWithdrawalConsequenceReducedToZero(balance sc.U128) WithdrawalConsequence {
	return WithdrawalConsequence{sc.NewVaryingData(WithdrawalConsequenceReducedToZero, balance)}
}

func NewWithdrawalConsequenceSuccess() WithdrawalConsequence {
	return WithdrawalConsequence{sc.NewVaryingData(WithdrawalConsequenceSuccess)}
}

func (wc WithdrawalConsequence) IntoResult(keepNonZero bool) (sc.U128, error) {
	switch wc.VaryingData[0] {
	case WithdrawalConsequenceBalanceLow:
		return sc.U128{}, primitives.NewDispatchErrorToken(primitives.NewTokenErrorFundsUnavailable())
	case WithdrawalConsequenceWouldDie:
		return sc.U128{}, primitives.NewDispatchErrorToken(primitives.NewTokenErrorOnlyProvider())
	case WithdrawalConsequenceUnknownAsset:
		return sc.U128{}, primitives.NewDispatchErrorToken(primitives.NewTokenErrorUnknownAsset())
	case WithdrawalConsequenceUnderflow:
		return sc.U128{}, primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorUnderflow())
	case WithdrawalConsequenceOverflow:
		return sc.U128{}, primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
	case WithdrawalConsequenceFrozen:
		return sc.U128{}, primitives.NewDispatchErrorToken(primitives.NewTokenErrorFrozen())
	case WithdrawalConsequenceReducedToZero:
		if keepNonZero {
			return sc.U128{}, primitives.NewDispatchErrorToken(primitives.NewTokenErrorNotExpendable())
		}
		return wc.VaryingData[1].(sc.U128), nil
	case WithdrawalConsequenceSuccess:
		return sc.U128{}, nil
	default:
		return sc.U128{}, primitives.NewDispatchErrorOther("invalid WithdrawalConsequence type")
	}
}
