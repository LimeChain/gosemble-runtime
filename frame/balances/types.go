package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/types"
)

type negativeImbalance struct {
	types.Balance
	totalIssuance support.StorageValue[sc.U128]
}

func newNegativeImbalance(balance types.Balance, totalIssuance support.StorageValue[sc.U128]) negativeImbalance {
	return negativeImbalance{balance, totalIssuance}
}

func (ni negativeImbalance) Drop() error {
	issuance, err := ni.totalIssuance.Get()
	if err != nil {
		return err
	}
	sub := sc.SaturatingSubU128(issuance, ni.Balance)

	ni.totalIssuance.Put(sub)
	return nil
}

type positiveImbalance struct {
	types.Balance
	totalIssuance support.StorageValue[sc.U128]
}

func newPositiveImbalance(balance types.Balance, totalIssuance support.StorageValue[sc.U128]) positiveImbalance {
	return positiveImbalance{balance, totalIssuance}
}

func (pi positiveImbalance) Drop() error {
	issuance, err := pi.totalIssuance.Get()
	if err != nil {
		return err
	}
	add := sc.SaturatingAddU128(issuance, pi.Balance)

	pi.totalIssuance.Put(add)
	return nil
}
