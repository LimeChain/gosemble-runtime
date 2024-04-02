package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Mutate an account to some new value, or delete it entirely with `None`. Will enforce
// `ExistentialDeposit` law, annulling the account as needed. This will do nothing if the
// result of `f` is an `Err`.
//
// It returns both the result from the closure, and an optional amount of dust
// which should be handled once it is known that all nested mutates that could affect
// storage items what the dust handler touches have completed.
//
// NOTE: Doesn't do any preparatory work for creating a new account, so should only be used
// when it is known that the account already exists.
//
// NOTE: LOW-LEVEL: This will not attempt to maintain total issuance. It is expected that
// the caller will do this.
func (m Module) tryMutateAccountNew(who primitives.AccountId, data primitives.AccountData) (primitives.Balance, error) {
	m.ensureUpgraded(who)

	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return primitives.Balance{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	if err := m.updateProviders(who, acc, data); err != nil {
		return primitives.Balance{}, err
	}

	dust := constants.Zero
	if data.Free.Lt(m.constants.ExistentialDeposit) && data.Reserved.Eq(constants.Zero) {
		dust = data.Free
	} else {
		if err := m.Config.StoredMap.TryMutateExistsNoClosure(who, data); err != nil {
			return primitives.Balance{}, err // todo make sure we are returning dispatchErr
		}
	}

	// Handle any steps needed after mutating an account.
	//
	// This includes DustRemoval unbalancing, in the case than the `new` account's total
	// balance is non-zero but below ED.
	//
	// Updates `maybe_account` to `Some` iff the account has sufficient balance.
	// Evaluates `maybe_dust`, which is `Some` containing the dust to be dropped, iff
	// some dust should be dropped.
	//
	// We should never be dropping if reserved is non-zero. Reserved being non-zero
	// should imply that we have a consumer ref, so this is economically safe.
	if acc.Data == primitives.DefaultAccountData() && data.Free.Gt(constants.Zero) {
		m.Config.StoredMap.DepositEvent(newEventEndowed(m.Index, who, data.Free))
	}

	if dust.Gt(constants.Zero) {
		m.Config.StoredMap.DepositEvent(newEventDustLost(m.Index, who, dust))
	}

	return dust, nil
}

// Ensure the account `who` is using the new logic.
//
// Returns `true` if the account did get upgraded, `false` if it didn't need upgrading.
func (m Module) ensureUpgraded(who primitives.AccountId) (bool, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return false, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	if acc.Data.Flags.IsNewLogic() {
		return false, nil
	}

	acc.Data.Flags = acc.Data.Flags.SetNewLogic()

	if !acc.Data.Reserved.Eq(constants.Zero) && acc.Data.Frozen.Eq(constants.Zero) {
		if acc.Providers == 0 {
			m.logger.Warnf("account with a non-zero reserve balance has no provider refs, account_id: '%s'.", string(who.Bytes()))
			acc.Data.Free = sc.Max128(acc.Data.Free, m.Config.ExistentialDeposit)

			if _, err := m.Config.StoredMap.IncProviders(who); err != nil {
				return false, err
			}
		}

		if err := m.Config.StoredMap.IncConsumersWithoutLimit(who); err != nil {
			return false, err
		}
	}

	if err = m.Config.StoredMap.TryMutateExistsNoClosure(who, acc.Data); err != nil {
		return false, err
	}

	m.Config.StoredMap.DepositEvent(newEventUpgraded(m.Index, who))
	return true, nil
}

func (m Module) updateProviders(who primitives.AccountId, acc primitives.AccountInfo, newState primitives.AccountData) error {
	didProvide := acc.Data.Free.Gte(m.constants.ExistentialDeposit) && acc.Providers > 0
	didConsume := acc.Data != primitives.DefaultAccountData() && (acc.Data.Reserved.Gte(constants.Zero) || acc.Data.Frozen.Gte(constants.Zero))
	doesProvide := newState.Free.Gte(m.constants.ExistentialDeposit)
	doesConsume := acc.Data.Reserved.Gte(constants.Zero) || acc.Data.Frozen.Gte(constants.Zero)

	if !didProvide && doesProvide {
		if _, err := m.Config.StoredMap.IncProviders(who); err != nil {
			return err
		}
	}
	if didConsume && !doesConsume {
		if err := m.Config.StoredMap.DecConsumers(who); err != nil {
			return err
		}
	}
	if !didConsume && doesConsume {
		if err := m.Config.StoredMap.IncConsumers(who); err != nil {
			return err
		}
	}
	if didProvide && !doesProvide {
		_, err := m.Config.StoredMap.DecProviders(who)
		if err != nil {
			if didConsume && !doesConsume {
				if err := m.Config.StoredMap.IncConsumers(who); err != nil {
					m.logger.Warnf("failed to increase consumers: %v", err)
				}
			}
			if !didConsume && doesConsume {
				if err := m.Config.StoredMap.DecConsumers(who); err != nil {
					m.logger.Warnf("failed to decrease consumers: %v", err)
				}
			}
			return err
		}
	}
	return nil
}
