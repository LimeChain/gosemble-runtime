package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	// "github.com/LimeChain/gosemble/frame/support"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// type balanceStore struct {
// 	eventStore         primitives.EventStore
// 	accountStore       primitives.StoredMap
// 	moduleIndex        sc.U8
// 	totalIssuance      support.StorageValue[sc.U128]
// 	existentialDeposit sc.U128
// }

// func newBalanceStore(eventStore primitives.EventStore, accountStore primitives.StoredMap, moduleIndex sc.U8, totalIssuance support.StorageValue[sc.U128], existientialDeposit sc.U128) balanceStore {
// 	return balanceStore{
// 		eventStore:         eventStore,
// 		accountStore:       accountStore,
// 		moduleIndex:        moduleIndex,
// 		totalIssuance:      totalIssuance,
// 		existentialDeposit: existientialDeposit,
// 	}
// }

func (m Module) tryMutateAccountNewCurrency(who primitives.AccountId, newAccData primitives.AccountData) (primitives.Balance, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return primitives.Balance{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	if err := m.updateProviders(who, acc, newAccData); err != nil {
		return primitives.Balance{}, err
	}

	dust := primitives.Balance(constants.Zero)
	if newAccData.Free.Lt(m.constants.ExistentialDeposit) && newAccData.Reserved.Eq(constants.Zero) {
		dust = newAccData.Free
	} else {
		// // todo
		// if _, err := m.Config.StoredMap.TryMutateExists(who, newAccData); err != nil {
		// 	return primitives.Balance{}, err
		// }
	}

	if (acc.Data == primitives.AccountData{}) {
		// if err := m.Config.StoredMap.DepositEvent(newEventEndowed(m.Index, who, newAccData.Free)); err != nil {
		// 	return primitives.Balance{}, err
		// }
	}

	if dust.Gt(constants.Zero) {
		// if err := m.Config.StoredMap.DepositEvent(newEventDustLost(m.Index, who, dust)); err != nil {
		// 	return primitives.Balance{}, err
		// }
	}

	return dust, nil
}

func (m Module) updateProvidersCurrency(who primitives.AccountId, acc primitives.AccountInfo, newState primitives.AccountData) error {
	// todo consumers?
	didProvide := acc.Data.Free.Gte(m.constants.ExistentialDeposit) && acc.Providers > 0
	doesProvide := newState.Free.Gte(m.constants.ExistentialDeposit)

	if !didProvide && doesProvide {
		if _, err := m.Config.StoredMap.IncProviders(who); err != nil {
			return err
		}
	}

	if didProvide && !doesProvide {
		if _, err := m.Config.StoredMap.DecProviders(who); err != nil {
			return err
		}
	}

	return nil
}

// if newState.Free.Gte(m.constants.ExistentialDeposit)
// dust := primitives.Balance(constants.Zero)
// if newState.Free.Eq(constants.Zero) || newState.Free.Gte(m.constants.ExistentialDeposit) || newState.Reserved.Gt(constants.Zero) {
// 	if err := m.Config.StoredMap.MutateAccountData(who, newState); err != nil {
// 		return err
// 	}
// } else {}

// dust := primitives.Balance{}
// if newState.Free.Lt(m.constants.ExistentialDeposit) && newState.Reserved.Eq(sc.NewU128(0)) {
// 	dust = newState.Free
// }

// todo when assert should we return error and stop ?:
// assert!(account.free.is_zero() || account.free >= ed || !account.reserved.is_zero());
// see https://github.com/LimeChain/polkadot-sdk/blob/ac7ab92efc12a9206261797484066277a3ff2139/substrate/frame/balances/src/lib.rs#L1010
// if dust.Eq(constants.Zero) {

// }

// func (m Module) HandleDust(who primitives.AccountId, dust Imbalance) error {
// 	if dust == nil
// 	if dust.Eq(constants.Zero) {
// 		return nil
// 	}

// 	issuance, err := m.totalIssuance.Get()
// 	if err != nil {
// 		return err
// 	}

// 	m.totalIssuance.Put(sc.SaturatingSubU128(issuance, dust))

// 	return m.negativeImbalance.Value.Drop()

// 	return nil
// }

// Withdraw withdraws `value` free balance from `who`, respecting existence requirements.
// Does not do anything if value is 0.
func (m Module) WithdrawNewCurrency(who primitives.AccountId, value primitives.Balance, reasons primitives.Reasons, liveness primitives.ExistenceRequirement) (primitives.Balance, error) {
	if value.Eq(constants.Zero) {
		return sc.U128{}, nil
	}

	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return sc.U128{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	newFree, err := sc.CheckedSubU128(acc.Data.Free, value)
	if err != nil {
		return sc.U128{}, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorInsufficientBalance),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	wouldBeDead := newFree.Lt(m.constants.ExistentialDeposit)
	wouldKill := wouldBeDead && acc.Data.Free.Gte(m.constants.ExistentialDeposit)
	if wouldKill && liveness == primitives.ExistenceRequirementKeepAlive {
		return sc.U128{}, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorKeepAlive),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if acc.Data.Frozen.Gt(newFree) {
		// if acc.Data.Frozen(reasons).Gt(newFree) { // todo
		return sc.U128{}, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorLiquidityRestrictions),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if newFree.Lt(m.constants.ExistentialDeposit) && newFree.Gt(constants.Zero) && acc.Data.Reserved.Eq(constants.Zero) {
		if newNegativeImbalance(newFree, m.storage.TotalIssuance).Drop(); err != nil {
			return sc.U128{}, err
		}

		m.Config.StoredMap.DepositEvent(newEventDustLost(m.Index, who, newFree))
	}

	m.Config.StoredMap.DepositEvent(newEventWithdraw(m.Index, who, value))

	// todo finish handleDust logic
	if _, err := m.tryMutateAccountNew(who, acc.Data); err != nil {
		return sc.U128{}, err
	}
	// if err := m.Config.StoredMap.MutateAccountData(who, acc.Data); err != nil {
	// 	return sc.U128{}, err
	// }

	return sc.U128{}, nil // todo return value
}

// type OnDropCredit = fungible::DecreaseIssuance<T::AccountId, Self>;
// type OnDropDebt = fungible::IncreaseIssuance<T::AccountId, Self>;

// func (m Module) HandleDust(who primitives.AccountId, newState primitives.Balance) error {
// 	acc, err := m.Config.StoredMap.Get(who)
// 	if err != nil {
// 		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
// 	}

// 	if dust.Eq(constants.Zero) {
// 		return nil
// 	}

// 	issuance, err := m.totalIssuance.Get()
// 	if err != nil {
// 		return err
// 	}

// 	m.totalIssuance.Put(sc.SaturatingSubU128(issuance, dust))

// 	return nil
// }

// todo eventually implement it vice versa:

// func (m Module) Drop(imbalance Imbalance) error {
// 	issuance, err := m.totalIssuance.Get()
// 	if err != nil {
// 		return err
// 	}

// 	switch i.imbalanceDirection {
// 	case positive:
// 		issuance = sc.SaturatingAddU128(issuance, i.Balance)
// 	case negative:
// 		issuance = sc.SaturatingSubU128(issuance, i.Balance)
// 	}

// 	i.store.Put(issuance)
// 	return nil
// }

// func (m Module) withdraw(who primitives.AccountId, value sc.U128, account *primitives.AccountData, reasons sc.U8, liveness primitives.ExistenceRequirement) (sc.Encodable, error) {
// 	newFreeAccount, err := sc.CheckedSubU128(account.Free, value)
// 	if err != nil {
// 		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 			Index:   m.Index,
// 			Err:     sc.U32(ErrorInsufficientBalance),
// 			Message: sc.NewOption[sc.Str](nil),
// 		})
// 	}

// 	wouldBeDead := (newFreeAccount.Add(account.Reserved)).Lt(m.constants.ExistentialDeposit)
// 	wouldKill := wouldBeDead && ((account.Free.Add(account.Reserved)).Gte(m.constants.ExistentialDeposit))

// 	if wouldKill && liveness == primitives.ExistenceRequirementKeepAlive {
// 		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 			Index:   m.Index,
// 			Err:     sc.U32(ErrorKeepAlive),
// 			Message: sc.NewOption[sc.Str](nil),
// 		})
// 	}

// 	// if err := bm.ensureCanWithdraw(who, value, primitives.Reasons(reasons), newFreeAccount); err != nil {
// 	// 	return nil, err
// 	// }
// 	if value.Eq(constants.Zero) {
// 		return nil, nil
// 	}

// 	accountInfo, err := m.Config.StoredMap.Get(who)
// 	if err != nil {
// 		return nil, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
// 	}

// 	minBalance := accountInfo.Data.Frozen(primitives.Reasons(reasons))
// 	if minBalance.Gt(newFreeAccount) {
// 		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 			Index:   m.Index,
// 			Err:     sc.U32(ErrorLiquidityRestrictions),
// 			Message: sc.NewOption[sc.Str](nil),
// 		})
// 	}

// 	account.Free = newFreeAccount

// 	m.Config.StoredMap.DepositEvent(newEventWithdraw(m.Index, who, value))
// 	return value, nil
// }

// ensureCanWithdraw checks that an account can withdraw from their balance given any existing withdraw restrictions.
// func (bm balanceStore) ensureCanWithdraw(who primitives.AccountId, amount sc.U128, reasons primitives.Reasons, newBalance sc.U128) error {
// 	if amount.Eq(constants.Zero) {
// 		return nil
// 	}

// 	accountInfo, err := bm.Config.StoredMap.Get(who)
// 	if err != nil {
// 		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
// 	}

// 	minBalance := accountInfo.Frozen(reasons)
// 	if minBalance.Gt(newBalance) {
// 		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 			Index:   bm.Index,
// 			Err:     sc.U32(ErrorLiquidityRestrictions),
// 			Message: sc.NewOption[sc.Str](nil),
// 		})
// 	}

// 	return nil
// }

// example api
// func (as accountStore) IncProviders(who primitives.AccountId) (primitives.IncRefStatus, error) {
// 	incRefStatus, err := as.storage.Account.Mutate(who, func(account *primitives.AccountInfo) (sc.Encodable, error) {
// 		incRefStatus := account.IncrementProviders()
// 		if incRefStatus == primitives.IncRefStatusCreated {
// 			if err := as.eventStore.DepositEvent(newEventNewAccount(as.moduleIndex, who)); err != nil {
// 				return incRefStatus, err
// 			}
// 		}

// 		return incRefStatus, nil
// 	})

// 	return incRefStatus.(primitives.IncRefStatus), err
// }
