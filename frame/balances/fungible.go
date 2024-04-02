package balances

import (
	// "reflect"

	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"

	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Transfer funds from one account into another.
//
// A transfer where the source and destination account are identical is treated as No-OP after
// checking the preconditions.
func (m Module) transfer(from primitives.AccountId, to primitives.AccountId, amount primitives.Balance, preservation balancestypes.Preservation) error {
	withdrawConsequence, err := m.canWithdraw(from, amount)
	if err != nil {
		return err
	}
	if err := withdrawConsequence.ToResult(preservation); err != nil {
		return err
	}

	depositConsequence, err := m.canDeposit(to, amount, false)
	if err != nil {
		return err
	}
	if err := depositConsequence.ToResult(); err != nil {
		return err
	}

	if reflect.DeepEqual(from, to) {
		return nil
	}

	// todo its possible that in substrate decreaseBalance and increaseBalance accept amount pointer and modify the value
	if _, err = m.decreaseBalance(from, amount, false, preservation, false); err != nil {
		return err
	}

	// This should never fail as we checked `can_deposit` earlier. But we do a best-effort anyway.
	m.increaseBalance(to, amount, false)

	m.Config.StoredMap.DepositEvent(newEventTransfer(m.Index, from, to, amount))
	return nil
}

// Returns `Success` if the balance of `who` may be decreased by `amount`, otherwise
// the consequence.
func (m Module) canWithdraw(who primitives.AccountId, amount primitives.Balance) (balancestypes.WithdrawConsequence, error) {
	if amount.Eq(constants.Zero) {
		return balancestypes.WithdrawConsequenceSuccess, nil
	}

	totalIssuance, err := m.storage.TotalIssuance.Get()
	if err != nil {
		return 0, err
	}

	if _, err := sc.CheckedSubU128(totalIssuance, amount); err != nil {
		return balancestypes.WithdrawConsequenceUnderflow, nil
	}

	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return 0, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	newFreeBalance, err := sc.CheckedSubU128(acc.Data.Free, amount)
	if err != nil {
		return balancestypes.WithdrawConsequenceBalanceLow, nil
	}

	liquid, err := m.reducibleBalance(who, balancestypes.PreservationExpendable, false)
	if err != nil {
		return balancestypes.WithdrawConsequenceBalanceLow, nil
	}

	if amount.Gt(liquid) {
		return balancestypes.WithdrawConsequenceFrozen, nil
	}

	// Provider restriction - total account balance cannot be reduced to zero if it cannot
	// sustain the loss of a provider reference.
	// NOTE: This assumes that the pallet is a provider (which is true). Is this ever changes,
	// then this will need to adapt accordingly.

	canDecProviders, err := m.Config.StoredMap.CanDecProviders(who)
	if err != nil {
		return 0, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	var withdrawConsequence balancestypes.WithdrawConsequence
	if newFreeBalance.Lt(m.constants.ExistentialDeposit) {
		if canDecProviders {
			withdrawConsequence = balancestypes.WithdrawConsequenceReducedToZero
		} else {
			return balancestypes.WithdrawConsequenceWouldDie, nil
		}
	} else {
		withdrawConsequence = balancestypes.WithdrawConsequenceSuccess
	}

	newTotalBalance := sc.SaturatingAddU128(newFreeBalance, acc.Data.Reserved)
	if newTotalBalance.Lt(acc.Data.Frozen) {
		return balancestypes.WithdrawConsequenceFrozen, nil
	}

	return withdrawConsequence, nil
}

// Returns `true` if the balance of `who` may be increased by `amount`.
//
// - `who`: The account of which the balance should be increased by `amount`.
// - `amount`: How much should the balance be increased?
// - `provenance`: Will `amount` be minted to deposit it into `account` or is it already in the system?
func (m Module) canDeposit(who primitives.AccountId, amount primitives.Balance, minted bool) (balancestypes.DepositConsequence, error) {
	if amount.Eq(constants.Zero) {
		return balancestypes.DepositConsequenceSuccess, nil
	}

	if minted {
		if totalIssuance, err := m.storage.TotalIssuance.Get(); err != nil {
			return 0, err
		} else if _, err := sc.CheckedAddU128(totalIssuance, amount); err != nil {
			return balancestypes.DepositConsequenceOverflow, nil // todo temp replace err with nil
		}
	}

	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return 0, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	newFree, err := sc.CheckedAddU128(acc.Data.Free, amount)
	if err != nil {
		return balancestypes.DepositConsequenceOverflow, nil
	} else if newFree.Lt(m.constants.ExistentialDeposit) {
		m.logger.Warn("line 148")
		return balancestypes.DepositConsequenceBelowMinimum, nil
	}

	if _, err := sc.CheckedAddU128(acc.Data.Reserved, newFree); err != nil {
		return balancestypes.DepositConsequenceOverflow, nil
	}

	// NOTE: We assume that we are a provider, so don't need to do any checks in the
	// case of account creation.
	return balancestypes.DepositConsequenceSuccess, nil
}

// Get the maximum amount that `who` can withdraw/transfer successfully based on whether the
// account should be kept alive (`preservation`) or whether we are willing to force the
// reduction and potentially go below user-level restrictions on the minimum amount of the account.
//
// Always less than or equal to [`Inspect::balance`].
func (m Module) reducibleBalance(who primitives.AccountId, preservation balancestypes.Preservation, force bool) (primitives.Balance, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		// Frozen balance applies to total. Anything on hold therefore gets discounted from the
		// limit given by the freezes.
		return primitives.Balance{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	untouchable := sc.NewU128(0)
	if !force {
		// Frozen balance applies to total. Anything on hold therefore gets discounted from the limit given by the freezes.
		untouchable = sc.SaturatingSubU128(acc.Data.Frozen, acc.Data.Reserved)
	}

	canDecProviders, err := m.Config.StoredMap.CanDecProviders(who)
	if err != nil {
		return primitives.Balance{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	// If we want to keep our provider ref..
	if preservation == balancestypes.PreservationPreserve ||
		// ..or we don't want the account to die and our provider ref is needed for it to live..
		(preservation == balancestypes.PreservationProtect && acc.Data.Free.Gt(constants.Zero) && acc.Providers == 1) ||
		// ..or we don't care about the account dying but our provider ref is required..
		(preservation == balancestypes.PreservationExpendable && acc.Data.Free.Gt(constants.Zero) && !canDecProviders) {
		// ..then the ED needed..
		untouchable = sc.Max128(untouchable, m.Config.ExistentialDeposit)
	}

	// Liquid balance is what is neither on hold nor frozen/required for provider.
	return sc.SaturatingSubU128(acc.Data.Free, untouchable), nil
}

func (m Module) decreaseBalance(who primitives.AccountId, amount primitives.Balance, exactPrecisiion bool, preservation balancestypes.Preservation, force bool) (primitives.Balance, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return constants.Zero, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	reducibleBalance, err := m.reducibleBalance(who, preservation, force)
	if err != nil {
		return constants.Zero, err
	}

	if exactPrecisiion {
		if amount.Gt(reducibleBalance) {
			return constants.Zero, primitives.NewDispatchErrorToken(primitives.NewTokenErrorNoFunds())
		}
	}

	if !exactPrecisiion {
		amount = sc.Min128(amount, reducibleBalance)
	}

	oldBalance := acc.Data.Free

	newBalance, err := sc.CheckedSubU128(oldBalance, amount)
	if err != nil {
		return constants.Zero, primitives.NewDispatchErrorToken(primitives.NewTokenErrorNoFunds())
	}

	dust, err := m.writeBalance(who, newBalance)
	if err != nil {
		return constants.Zero, err
	}

	if dust.Gt(constants.Zero) {
		// todo handle_dust
		totalIssuance, err := m.storage.TotalIssuance.Get()
		if err != nil {
			return constants.Zero, err
		}

		m.storage.TotalIssuance.Put(sc.SaturatingSubU128(totalIssuance, dust))
	}

	return sc.SaturatingSubU128(oldBalance, newBalance), nil
}

func (m Module) increaseBalance(who primitives.AccountId, amount primitives.Balance, exactPrecisiion bool) (primitives.Balance, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return constants.Zero, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	oldBalance := acc.Data.Free
	newBalance := constants.Zero
	if exactPrecisiion {
		newBalance, err = sc.CheckedAddU128(oldBalance, amount)
		if err != nil {
			return constants.Zero, primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
		}
	}

	if !exactPrecisiion {
		newBalance = sc.SaturatingAddU128(oldBalance, amount)
	}

	if newBalance.Lt(m.constants.ExistentialDeposit) {
		if exactPrecisiion {
			m.logger.Warnf("line 261")
			return constants.Zero, primitives.NewDispatchErrorToken(primitives.NewTokenErrorBelowMinimum())
		}

		return constants.Zero, nil
	}

	if newBalance.Eq(oldBalance) {
		return constants.Zero, nil
	}

	dust, err := m.writeBalance(who, newBalance)
	if err != nil {
		return constants.Zero, err
	}

	if dust.Gt(constants.Zero) {
		// todo handle_dust
		totalIssuance, err := m.storage.TotalIssuance.Get()
		if err != nil {
			return constants.Zero, err
		}

		m.storage.TotalIssuance.Put(sc.SaturatingSubU128(totalIssuance, dust))
	}

	return sc.SaturatingSubU128(newBalance, oldBalance), nil
}

// returns dust, err
func (m Module) writeBalance(who primitives.AccountId, amount primitives.Balance) (primitives.Balance, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return constants.Zero, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	maxReduction, err := m.reducibleBalance(who, balancestypes.PreservationExpendable, true)
	if err != nil {
		return constants.Zero, err
	}

	if sc.SaturatingSubU128(acc.Data.Free, maxReduction).Gt(maxReduction) {
		return constants.Zero, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorInsufficientBalance),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	acc.Data.Free = amount

	dust, err := m.tryMutateAccountNew(who, acc.Data)
	if err != nil {
		return constants.Zero, err
	}

	return dust, nil
}
