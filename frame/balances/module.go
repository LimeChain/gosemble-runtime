package balances

import (
	"encoding/hex"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/frame/balances/types"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"reflect"
)

const (
	functionTransferAllowDeath       sc.U8 = iota
	functionForceTransfer            sc.U8 = 2
	functionTransferKeepAlive        sc.U8 = 3
	functionTransferAll              sc.U8 = 4
	functionForceUnreserve           sc.U8 = 5
	functionForceUpgradeAccounts     sc.U8 = 6
	functionForceSetBalance          sc.U8 = 8
	functionForceAdjustTotalIssuance sc.U8 = 9
)

const (
	name = sc.Str("Balances")
)

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	Index       sc.U8
	Config      *Config
	constants   *consts
	storage     *storage
	functions   map[sc.U8]primitives.Call
	mdGenerator *primitives.MetadataTypeGenerator
	logger      log.Logger
}

func New(index sc.U8, config *Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.Logger) Module {
	constants := newConstants(config.DbWeight, config.MaxLocks, config.MaxReserves, config.ExistentialDeposit)
	storage := newStorage()

	module := Module{
		Index:       index,
		Config:      config,
		constants:   constants,
		storage:     storage,
		mdGenerator: mdGenerator,
		logger:      logger,
	}
	functions := make(map[sc.U8]primitives.Call)
	functions[functionTransferAllowDeath] = newCallTransferAllowDeath(index, functionTransferAllowDeath, module)
	functions[functionForceTransfer] = newCallForceTransfer(index, functionForceTransfer, module)
	functions[functionTransferKeepAlive] = newCallTransferKeepAlive(index, functionTransferKeepAlive, module)
	functions[functionTransferAll] = newCallTransferAll(index, functionTransferAll, module)
	functions[functionForceUnreserve] = newCallForceUnreserve(index, functionForceUnreserve, module)
	functions[functionForceUpgradeAccounts] = newCallUpgradeAccounts(index, functionForceUpgradeAccounts, module)
	functions[functionForceSetBalance] = newCallForceSetBalance(index, functionForceSetBalance, module)
	functions[functionForceAdjustTotalIssuance] = newCallForceAdjustTotalIssuance(index, functionForceAdjustTotalIssuance, config.StoredMap, storage)

	module.functions = functions

	return module
}

func (m Module) GetIndex() sc.U8 {
	return m.Index
}

func (m Module) name() sc.Str {
	return name
}

func (m Module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m Module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m Module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

// DepositIntoExisting deposits `value` into the free balance of an existing target account `who`.
// If `value` is 0, it does nothing.
func (m Module) DepositIntoExisting(who primitives.AccountId, value sc.U128) (primitives.Balance, error) {
	if value.Eq(constants.Zero) {
		return sc.NewU128(0), nil
	}

	result, err := m.tryMutateAccountHandlingDust(
		who,
		func(accountData *primitives.AccountData, isNew bool) (sc.Encodable, error) {
			return m.deposit(who, accountData, isNew, value)
		},
	)
	if err != nil {
		return primitives.Balance{}, err
	}

	return result.(primitives.Balance), nil
}

func (m Module) Withdraw(who primitives.AccountId, value sc.U128, reasons sc.U8, liveness primitives.ExistenceRequirement) (primitives.Balance, error) {
	if value.Eq(constants.Zero) {
		return sc.NewU128(0), nil
	}

	result, err := m.tryMutateAccountHandlingDust(
		who,
		func(accountData *primitives.AccountData, isNew bool) (sc.Encodable, error) {
			return m.withdraw(who, value, accountData, reasons, liveness)
		},
	)

	if err != nil {
		return primitives.Balance{}, err
	}

	return result.(primitives.Balance), nil
}

func (m Module) deposit(who primitives.AccountId, account *primitives.AccountData, isNew bool, value sc.U128) (sc.Encodable, error) {
	if isNew {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorDeadAccount),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	free, err := sc.CheckedAddU128(account.Free, value)
	if err != nil {
		return nil, primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
	}
	account.Free = free

	m.Config.StoredMap.DepositEvent(newEventDeposit(m.Index, who, value))

	return value, nil
}

func (m Module) withdraw(who primitives.AccountId, value sc.U128, account *primitives.AccountData, reasons sc.U8, liveness primitives.ExistenceRequirement) (sc.Encodable, error) {
	newFreeAccount, err := sc.CheckedSubU128(account.Free, value)
	if err != nil {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorInsufficientBalance),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	existentialDeposit := m.constants.ExistentialDeposit

	wouldBeDead := newFreeAccount.Lt(existentialDeposit)
	wouldKill := wouldBeDead && account.Free.Gte(existentialDeposit)
	if !(liveness == primitives.ExistenceRequirementAllowDeath || !wouldKill) {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorExpendability),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if err := m.ensureCanWithdraw(who, value, primitives.Reasons(reasons), newFreeAccount); err != nil {
		return nil, err
	}

	account.Free = newFreeAccount
	m.Config.StoredMap.DepositEvent(newEventWithdraw(m.Index, who, value))
	return value, nil
}

// ensureCanWithdraw checks that an account can withdraw from their balance given any existing withdraw restrictions.
func (m Module) ensureCanWithdraw(who primitives.AccountId, amount sc.U128, _reasons primitives.Reasons, newBalance sc.U128) error {
	if amount.Eq(constants.Zero) {
		return nil
	}

	accountInfo, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	minBalance := accountInfo.Data.Frozen
	if minBalance.Gt(newBalance) {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorLiquidityRestrictions),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	return nil
}

func (m Module) ensureUpgraded(who primitives.AccountId) (bool, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return false, err
	}

	if acc.Data.Flags.IsNewLogic() {
		return false, nil
	}
	acc.Data.Flags = acc.Data.Flags.SetNewLogic()
	if !acc.Data.Reserved.Eq(constants.Zero) && acc.Data.Frozen.Eq(constants.Zero) {
		if acc.Providers == 0 {
			m.logger.Warnf("account with a non-zero reserve balance has no provider refs, acc_id [%s]", hex.EncodeToString(who.Bytes()))
			acc.Data.Free = sc.Max128(acc.Data.Free, m.constants.ExistentialDeposit)
			_, err := m.Config.StoredMap.IncProviders(who)
			if err != nil {
				return false, err
			}
		}

		err := m.Config.StoredMap.IncConsumersWithoutLimit(who)
		if err != nil {
			return false, err
		}
	}
	_, err = m.Config.StoredMap.TryMutateExists(who, func(target *primitives.AccountData) (sc.Encodable, error) {
		updateAccount(target, acc.Data)
		return nil, nil
	})
	if err != nil {
		return false, err
	}

	m.Config.StoredMap.DepositEvent(newEventUpgraded(m.Index, who))

	return true, nil
}

func (m Module) transfer(from primitives.AccountId, to primitives.AccountId, value sc.U128, preservation types.Preservation) error {
	withdrawalConsequence, err := m.canWithdraw(from, value)
	if err != nil {
		return err
	}
	_, err = withdrawalConsequence.IntoResult(preservation != types.PreservationExpendable)
	if err != nil {
		return err
	}
	depositConsequence, err := m.canDeposit(to, value, false)
	if err != nil {
		return err
	}
	err = depositConsequence.IntoResult()
	if err != nil {
		return err
	}

	if reflect.DeepEqual(from, to) {
		return nil
	}

	_, err = m.decreaseBalance(from, value, types.PrecisionBestEffort, preservation, types.FortitudePolite)
	if err != nil {
		return err
	}

	// This should never fail as we checked `can_deposit` earlier. But we do a best-effort
	// anyway.
	_, err = m.increaseBalance(to, value, types.PrecisionBestEffort)
	if err != nil {
		return err
	}

	m.Config.StoredMap.DepositEvent(newEventTransfer(m.Index, from, to, value))

	return nil
}

func (m Module) increaseBalance(who primitives.AccountId, amount sc.U128, precision types.Precision) (sc.U128, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return sc.U128{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	oldBalance := acc.Data.Free

	var newBalance sc.U128
	if precision == types.PrecisionBestEffort {
		newBalance = sc.SaturatingAddU128(oldBalance, amount)
	} else {
		newBalance, err = sc.CheckedAddU128(oldBalance, amount)
		if err != nil {
			return sc.U128{}, primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
		}
	}

	if newBalance.Lt(m.constants.ExistentialDeposit) {
		if precision == types.PrecisionBestEffort {
			return constants.Zero, nil
		} else {
			return sc.U128{}, primitives.NewDispatchErrorToken(primitives.NewTokenErrorBelowMinimum())
		}
	}

	if newBalance.Eq(oldBalance) {
		return constants.Zero, nil
	}

	dust, err := m.writeBalance(who, newBalance)
	if err != nil {
		return sc.U128{}, err
	}

	if dust.HasValue {
		err := m.handleDust(dust.Value)
		if err != nil {
			return sc.U128{}, err
		}
	}

	return sc.SaturatingSubU128(newBalance, oldBalance), nil
}

func (m Module) decreaseBalance(who primitives.AccountId, value sc.U128, precision types.Precision, preservation types.Preservation, fortitude types.Fortitude) (sc.U128, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return sc.U128{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	oldBalance := acc.Data.Free

	reducible, err := m.reducibleBalance(who, preservation, fortitude)
	if err != nil {
		return sc.U128{}, err
	}
	if precision == types.PrecisionBestEffort {
		value = sc.Min128(value, reducible)
	} else if precision == types.PrecisionExact {
		if value.Gt(reducible) {
			return sc.U128{}, primitives.NewDispatchErrorToken(primitives.NewTokenErrorFundsUnavailable())
		}
	}

	newBalance, err := sc.CheckedSubU128(oldBalance, value)
	if err != nil {
		return sc.U128{}, primitives.NewDispatchErrorToken(primitives.NewTokenErrorFundsUnavailable())
	}

	maybeDust, err := m.writeBalance(who, newBalance)
	if err != nil {
		return sc.U128{}, err
	}
	if maybeDust.HasValue {
		err := m.handleDust(maybeDust.Value)
		if err != nil {
			return sc.U128{}, err
		}
	}

	return sc.SaturatingSubU128(oldBalance, newBalance), nil

}

// handleRawDust creates some dust and handle it with [`Unbalanced::handle_dust`]. This is an unbalanced
//
//	operation, so it must only be used when an account is modified in a raw fashion, outside
//
// the entire fungibles API. The `amount` is capped at [`Inspect::minimum_balance()`] - 1`.
//
// This should not be reimplemented.
func (m Module) handleRawDust(dust sc.U128) error {
	return m.handleDust(sc.Min128(dust, sc.SaturatingSubU128(m.Config.ExistentialDeposit, constants.One)))
}

func (m Module) handleDust(dust sc.U128) error {
	// TODO: handle dust
	return nil
}

func (m Module) writeBalance(who primitives.AccountId, amount sc.U128) (sc.Option[sc.U128], error) {
	maxReduction, err := m.reducibleBalance(who, types.PreservationExpendable, types.FortitudeForce)
	if err != nil {
		return sc.Option[sc.U128]{}, err
	}

	result, err := m.tryMutateAccount(who, func(accountData *primitives.AccountData, bool bool) (sc.Encodable, error) {
		reduction := sc.SaturatingSubU128(accountData.Free, amount)
		if reduction.Gt(maxReduction) {
			return nil, primitives.NewDispatchErrorModule(
				primitives.CustomModuleError{
					Index:   m.Index,
					Err:     sc.U32(ErrorInsufficientBalance),
					Message: sc.NewOption[sc.Str](nil),
				})
		}
		accountData.Free = amount
		return nil, nil
	})
	if err != nil {
		return sc.Option[sc.U128]{}, err
	}

	resultValue := result.(sc.VaryingData)
	return resultValue[1].(sc.Option[sc.U128]), nil
}

func (m Module) tryMutateAccountHandlingDust(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error) {
	result, err := m.tryMutateAccount(who, f)
	if err != nil {
		return result, err
	}

	resultValue := result.(sc.VaryingData)
	maybeDust, ok := resultValue[1].(sc.Option[sc.U128])
	if !ok {
		return nil, primitives.NewDispatchErrorOther("could not cast dust in mutateAccountHandlingDust")
	}

	if maybeDust.HasValue {
		err := m.handleRawDust(maybeDust.Value)
		if err != nil {
			return nil, err
		}
	}

	return resultValue[0], nil
}

func (m Module) mutateAccountHandlingDust(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error) {
	result, err := m.tryMutateAccount(who, f)
	if err != nil {
		return result, err
	}

	resultValue := result.(sc.VaryingData)
	maybeDust, ok := resultValue[1].(sc.Option[sc.U128])
	if !ok {
		return nil, primitives.NewDispatchErrorOther("could not cast dust in mutateAccountHandlingDust")
	}

	if maybeDust.HasValue {
		err := m.handleRawDust(maybeDust.Value)
		if err != nil {
			return nil, err
		}
	}

	return resultValue[0], nil
}

func (m Module) tryMutateAccount(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error) {
	_, err := m.ensureUpgraded(who)
	if err != nil {
		return nil, nil
	}

	result, err := m.Config.StoredMap.TryMutateExists(who, func(maybeAccount *primitives.AccountData) (sc.Encodable, error) {
		return m.mutateAccount(who, maybeAccount, f)
	})
	if err != nil {
		return result, err
	}

	resultValue := result.(sc.VaryingData)
	maybeEndowed := resultValue[0].(sc.Option[primitives.Balance])
	if maybeEndowed.HasValue {
		m.Config.StoredMap.DepositEvent(newEventEndowed(m.Index, who, maybeEndowed.Value))
	}

	maybeDust := resultValue[1].(sc.Option[primitives.Balance])
	if maybeDust.HasValue {
		m.Config.StoredMap.DepositEvent(newEventDustLost(m.Index, who, maybeDust.Value))
	}

	return sc.NewVaryingData(resultValue[2], maybeDust), nil
}

func (m Module) mutateAccount(who primitives.AccountId, maybeAccount *primitives.AccountData, f func(who *primitives.AccountData, _ bool) (sc.Encodable, error)) (sc.Encodable, error) {
	data := primitives.DefaultAccountData()
	account := &data
	isNew := true
	if !reflect.DeepEqual(*maybeAccount, primitives.DefaultAccountData()) {
		account = maybeAccount
		isNew = false
	}

	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return primitives.Balance{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	didProvide := account.Free.Gte(m.constants.ExistentialDeposit) && acc.Providers > 0
	didConsume := !isNew && (!account.Reserved.Eq(constants.Zero) || !account.Frozen.Eq(constants.Zero))

	result, err := f(account, isNew)
	if err != nil {
		return result, err
	}

	doesProvide := account.Free.Gte(m.constants.ExistentialDeposit)
	doesConsume := !account.Reserved.Eq(constants.Zero) || !account.Frozen.Eq(constants.Zero)

	if !didProvide && doesProvide {
		_, err = m.Config.StoredMap.IncProviders(who)
		if err != nil {
			return nil, err
		}
	}
	if didConsume && !doesConsume {
		err = m.Config.StoredMap.DecConsumers(who)
		if err != nil {
			return nil, err
		}
	}
	if !didConsume && doesConsume {
		err = m.Config.StoredMap.IncConsumers(who)
		if err != nil {
			return nil, err
		}
	}
	if didProvide && !doesProvide {
		// This could reap the account so must go last.
		_, err = m.Config.StoredMap.DecProviders(who)
		if err != nil {
			if didConsume && !doesConsume {
				err := m.Config.StoredMap.IncConsumers(who)
				if err != nil {
					m.logger.Criticalf("defensive: [%s]", err.Error())
				}
			}
			if !didConsume && doesConsume {
				err := m.Config.StoredMap.DecConsumers(who)
				if err != nil {
					return nil, err
				}
			}
			return nil, err
		}
	}

	maybeEndowed := sc.NewOption[primitives.Balance](nil)
	if isNew {
		maybeEndowed = sc.NewOption[primitives.Balance](account.Free)
	}

	maybeDust := sc.NewOption[primitives.Balance](nil)
	if account.Free.Lt(m.constants.ExistentialDeposit) && account.Reserved.Eq(constants.Zero) {
		if !account.Free.Eq(constants.Zero) {
			maybeDust = sc.NewOption[primitives.Balance](account.Free)
		}
	} else {
		if !(account.Free.Eq(constants.Zero) || account.Free.Gte(m.constants.ExistentialDeposit) || account.Reserved.Eq(constants.Zero)) {
			m.logger.Criticalf("failed to assert maybe dust")
		}
		maybeAccount.Free = account.Free
		maybeAccount.Reserved = account.Reserved
		maybeAccount.Frozen = account.Frozen
		maybeAccount.Flags = account.Flags
	}

	return sc.NewVaryingData(maybeEndowed, maybeDust, result), nil
}

func (m Module) canWithdraw(who primitives.AccountId, value sc.U128) (types.WithdrawalConsequence, error) {
	if value.Eq(constants.Zero) {
		return types.NewWithdrawalConsequenceSuccess(), nil
	}

	totalIssuance, err := m.storage.TotalIssuance.Get()
	if err != nil {
		return types.WithdrawalConsequence{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	if _, err := sc.CheckedSubU128(totalIssuance, value); err != nil {
		return types.NewWithdrawalConsequenceUnderflow(), nil
	}

	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return types.WithdrawalConsequence{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	newFreeBalance, err := sc.CheckedSubU128(acc.Data.Free, value)
	if err != nil {
		return types.NewWithdrawalConsequenceBalanceLow(), nil
	}

	liquid, err := m.reducibleBalance(who, types.PreservationExpendable, types.FortitudePolite)
	if err != nil {
		return types.WithdrawalConsequence{}, err
	}

	if value.Gt(liquid) {
		return types.NewWithdrawalConsequenceFrozen(), nil
	}

	// Provider restriction - total account balance cannot be reduced to zero if it cannot
	// sustain the loss of a provider reference.
	// NOTE: This assumes that the pallet is a provider (which is true). Is this ever changes,
	// then this will need to adapt accordingly.

	canDecProviders, err := m.Config.StoredMap.CanDecProviders(who)
	if err != nil {
		return types.WithdrawalConsequence{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	var success types.WithdrawalConsequence
	if newFreeBalance.Lt(m.constants.ExistentialDeposit) {
		if canDecProviders {
			success = types.NewWithdrawalConsequenceReducedToZero(newFreeBalance)
		} else {
			return types.NewWithdrawalConsequenceWouldDie(), nil
		}
	} else {
		success = types.NewWithdrawalConsequenceSuccess()
	}

	newTotalBalance := sc.SaturatingAddU128(newFreeBalance, acc.Data.Reserved)
	if newTotalBalance.Lt(acc.Data.Frozen) {
		return types.NewWithdrawalConsequenceFrozen(), nil
	}

	return success, nil
}

// Returns `true` if the balance of `who` may be increased by `amount`.
//
// - `who`: The account of which the balance should be increased by `amount`.
// - `amount`: How much should the balance be increased?
// - `provenance`: Will `amount` be minted to deposit it into `account` or is it already in the system?
func (m Module) canDeposit(who primitives.AccountId, amount primitives.Balance, minted bool) (types.DepositConsequence, error) {
	if amount.Eq(constants.Zero) {
		return types.NewDepositConsequenceSuccess(), nil
	}

	if minted {
		if totalIssuance, err := m.storage.TotalIssuance.Get(); err != nil {
			return types.DepositConsequence{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		} else if _, err := sc.CheckedAddU128(totalIssuance, amount); err != nil {
			return types.NewDepositConsequenceOverflow(), nil
		}
	}

	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return types.DepositConsequence{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	newFree, err := sc.CheckedAddU128(acc.Data.Free, amount)
	if err != nil {
		return types.NewDepositConsequenceOverflow(), nil
	}
	if newFree.Lt(m.constants.ExistentialDeposit) {
		return types.NewDepositConsequenceBelowMinimum(), nil
	}

	if _, err := sc.CheckedAddU128(acc.Data.Reserved, newFree); err != nil {
		return types.NewDepositConsequenceOverflow(), nil
	}

	// NOTE: We assume that we are a provider, so don't need to do any checks in the
	// case of account creation.
	return types.NewDepositConsequenceSuccess(), nil
}

// Get the maximum amount that `who` can withdraw/transfer successfully based on whether the
// account should be kept alive (`preservation`) or whether we are willing to force the
// reduction and potentially go below user-level restrictions on the minimum amount of the account.
//
// Always less than or equal to [`Inspect::balance`].
func (m Module) reducibleBalance(who primitives.AccountId, preservation types.Preservation, force types.Fortitude) (primitives.Balance, error) {
	acc, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return primitives.Balance{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	untouchable := sc.NewU128(0)
	if force == types.FortitudePolite {
		// Frozen balance applies to total. Anything on hold therefore gets discounted from the limit given by the freezes.
		untouchable = sc.SaturatingSubU128(acc.Data.Frozen, acc.Data.Reserved)
	}

	canDecProviders, err := m.Config.StoredMap.CanDecProviders(who)
	if err != nil {
		return primitives.Balance{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	// If we want to keep our provider ref
	if preservation == types.PreservationPreserve ||
		// ..or we don't want the account to die and our provider ref is needed for it to live..
		(preservation == types.PreservationProtect && !acc.Data.Free.Eq(constants.Zero) && acc.Providers == 1) ||
		// ..or we don't care about the account dying but our provider ref is required..
		(preservation == types.PreservationExpendable && !acc.Data.Free.Eq(constants.Zero) && !canDecProviders) {
		// ..then the ED needed..
		untouchable = sc.Max128(untouchable, m.Config.ExistentialDeposit)
	}

	// Liquid balance is what is neither on hold nor frozen/required for provider.
	return sc.SaturatingSubU128(acc.Data.Free, untouchable), nil
}

func (m Module) unreserve(who primitives.AccountId, value sc.U128) (sc.U128, error) {
	if value.Eq(constants.Zero) {
		return constants.Zero, nil
	}

	account, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return sc.U128{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	totalBalance := account.Data.Total()
	if totalBalance.Eq(constants.Zero) {
		return value, nil
	}

	result, err := m.mutateAccountHandlingDust(who, func(accountData *primitives.AccountData, bool bool) (sc.Encodable, error) {
		return removeReserveAndFree(accountData, value), nil
	})
	if err != nil {
		return sc.U128{}, err
	}
	actual := result.(primitives.Balance)
	m.Config.StoredMap.DepositEvent(newEventUnreserved(m.Index, who, actual))

	return value.Sub(actual), nil
}

// removeReserveAndFree frees reserved value from the account.
func removeReserveAndFree(account *primitives.AccountData, value sc.U128) primitives.Balance {
	actual := sc.Min128(account.Reserved, value)
	account.Reserved = account.Reserved.Sub(actual)

	account.Free = sc.SaturatingAddU128(account.Free, actual)

	return actual
}

func updateAccount(account *primitives.AccountData, data primitives.AccountData) {
	account.Free = data.Free
	account.Reserved = data.Reserved
	account.Frozen = data.Frozen
	account.Flags = data.Flags
}
