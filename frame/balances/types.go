package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/types"
)

// type accountMutator interface {
// 	ensureCanWithdraw(who types.AccountId, amount sc.U128, reasons types.Reasons, newBalance sc.U128) error
// 	tryMutateAccountWithDust(who types.AccountId, f func(who *types.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error)
// 	tryMutateAccount(who types.AccountId, f func(who *types.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error)
// 	ensureUpgraded(who types.AccountId) (bool, error)
// }

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

type dustCleaner struct {
	moduleIndex       sc.U8
	accountId         types.AccountId
	negativeImbalance sc.Option[negativeImbalance]
	eventDepositor    types.EventDepositor
}

func newDustCleaner(moduleId sc.U8, accountId types.AccountId, negativeImbalance sc.Option[negativeImbalance], eventDepositor types.EventDepositor) dustCleaner {
	return dustCleaner{
		moduleIndex:       moduleId,
		accountId:         accountId,
		negativeImbalance: negativeImbalance,
		eventDepositor:    eventDepositor,
	}
}

func (dcv dustCleaner) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		dcv.accountId,
		dcv.negativeImbalance,
	)
}

func (dcv dustCleaner) Bytes() []byte {
	return sc.EncodedBytes(dcv)
}

func (dcv dustCleaner) Drop() error {
	if dcv.negativeImbalance.HasValue {
		dcv.eventDepositor.DepositEvent(newEventDustLost(dcv.moduleIndex, dcv.accountId, dcv.negativeImbalance.Value.Balance))
		return dcv.negativeImbalance.Value.Drop()
	}

	return nil
}
