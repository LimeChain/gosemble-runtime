package balances

import (
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"

	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// todo remove??
type accountMutator interface {
	ensureCanWithdraw(who primitives.AccountId, amount sc.U128, reasons primitives.Reasons, newBalance sc.U128) error
	tryMutateAccountWithDust(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error)
	tryMutateAccount(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error)
	ensureUpgraded(who primitives.AccountId) (bool, error)
	transfer(from primitives.AccountId, to primitives.AccountId, amount sc.U128, preservation balancestypes.Preservation) error
}

// ensureCanWithdraw checks that an account can withdraw from their balance given any existing withdraw restrictions.
func (m Module) ensureCanWithdraw(who primitives.AccountId, amount sc.U128, reasons primitives.Reasons, newBalance sc.U128) error {
	if amount.Eq(constants.Zero) {
		return nil
	}

	accountInfo, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	minBalance := accountInfo.Data.Frozen //todo
	if minBalance.Gt(newBalance) {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorLiquidityRestrictions),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	return nil
}

// tryMutateAccount mutates an account based on argument `f`. Does not change total issuance.
// Does not do anything if `f` returns an error.
func (m Module) tryMutateAccount(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error) {
	result, err := m.tryMutateAccountWithDust(who, f)
	if err != nil {
		return result, err
	}

	r := result.(sc.VaryingData)

	dustCleaner := r[1].(dustCleaner)
	dustCleaner.Drop()

	return r[0].(sc.Encodable), nil
}

func (m Module) tryMutateAccountWithDust(who primitives.AccountId, f func(who *primitives.AccountData, _ bool) (sc.Encodable, error)) (sc.Encodable, error) {
	result, err := m.Config.StoredMap.TryMutateExists(
		who,
		func(maybeAccount *primitives.AccountData) (sc.Encodable, error) {
			return m.mutateAccount(maybeAccount, f)
		},
	)
	if err != nil {
		return result, err
	}

	resultValue := result.(sc.VaryingData)
	maybeEndowed := resultValue[0].(sc.Option[primitives.Balance])
	if maybeEndowed.HasValue {
		m.Config.StoredMap.DepositEvent(newEventEndowed(m.Index, who, maybeEndowed.Value))
	}

	maybeDust := resultValue[1].(sc.Option[negativeImbalance])
	dustCleaner := newDustCleaner(m.Index, who, maybeDust, m.Config.StoredMap)

	r := sc.NewVaryingData(resultValue[2], dustCleaner)
	return r, nil
}

func (m Module) mutateAccount(maybeAccount *primitives.AccountData, f func(who *primitives.AccountData, _ bool) (sc.Encodable, error)) (sc.Encodable, error) {
	defaultAcc := primitives.DefaultAccountData()
	account := &defaultAcc
	isNew := true
	if !reflect.DeepEqual(*maybeAccount, defaultAcc) {
		account = maybeAccount
		isNew = false
	}

	result, err := f(account, isNew)
	if err != nil {
		return result, err
	}

	maybeEndowed := sc.NewOption[primitives.Balance](nil)
	if isNew {
		maybeEndowed = sc.NewOption[primitives.Balance](account.Free)
	}
	maybeAccountWithDust, imbalance := m.postMutation(*account)
	if !maybeAccountWithDust.HasValue {
		*maybeAccount = primitives.DefaultAccountData()
	} else {
		maybeAccount.Free = maybeAccountWithDust.Value.Free
		maybeAccount.Frozen = maybeAccountWithDust.Value.Frozen
		maybeAccount.Reserved = maybeAccountWithDust.Value.Reserved
		maybeAccount.Flags = maybeAccountWithDust.Value.Flags
	}

	r := sc.NewVaryingData(maybeEndowed, imbalance, result)

	return r, nil
}

func (m Module) postMutation(new primitives.AccountData) (sc.Option[primitives.AccountData], sc.Option[negativeImbalance]) {
	total := new.Total()

	if total.Lt(m.constants.ExistentialDeposit) {
		if total.Eq(constants.Zero) {
			return sc.NewOption[primitives.AccountData](nil), sc.NewOption[negativeImbalance](nil)
		} else {
			return sc.NewOption[primitives.AccountData](nil), sc.NewOption[negativeImbalance](newNegativeImbalance(total, m.storage.TotalIssuance))
		}
	}

	return sc.NewOption[primitives.AccountData](new), sc.NewOption[negativeImbalance](nil)
}

func (m Module) withdraw(who primitives.AccountId, amount sc.U128, account *primitives.AccountData, reasons sc.U8, preservation balancestypes.Preservation, force bool) (sc.Encodable, error) {
	newFreeAccount, err := sc.CheckedSubU128(account.Free, amount)
	if err != nil {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorInsufficientBalance),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	existentialDeposit := m.constants.ExistentialDeposit

	wouldBeDead := (newFreeAccount.Add(account.Reserved)).Lt(existentialDeposit)
	wouldKill := wouldBeDead && ((account.Free.Add(account.Reserved)).Gte(existentialDeposit))
	m.logger.Warnf("wouldBeDead: %v, wouldKill: %v", wouldBeDead, wouldKill)
	if !(preservation == balancestypes.PreservationExpendable || !wouldKill) {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorKeepAlive),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if err := m.ensureCanWithdraw(who, amount, primitives.Reasons(reasons), newFreeAccount); err != nil {
		return nil, err
	}

	account.Free = newFreeAccount

	m.Config.StoredMap.DepositEvent(newEventWithdraw(m.Index, who, amount))
	return amount, nil
}

func (m Module) deposit(who primitives.AccountId, account *primitives.AccountData, isNew bool, amount sc.U128) (sc.Encodable, error) {
	if isNew {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorDeadAccount),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	free, err := sc.CheckedAddU128(account.Free, amount)
	if err != nil {
		return nil, primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
	}
	account.Free = free

	m.Config.StoredMap.DepositEvent(newEventDeposit(m.Index, who, amount))

	return amount, nil
}

// sanityChecks checks the following:
// `fromAccount` has sufficient balance
// `toAccount` balance does not overflow
// `toAccount` total balance is more than the existential deposit
// `fromAccount` can withdraw `value`
// the existence requirements for `fromAccount`
// Updates the balances of `fromAccount` and `toAccount`.
func (m Module) sanityChecks(from primitives.AccountId, fromAccount *primitives.AccountData, toAccount *primitives.AccountData, amount sc.U128, preservation balancestypes.Preservation) error {
	fromFree, err := sc.CheckedSubU128(fromAccount.Free, amount)
	if err != nil {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorInsufficientBalance),
			Message: sc.NewOption[sc.Str](nil),
		})
	}
	fromAccount.Free = fromFree

	toFree, err := sc.CheckedAddU128(toAccount.Free, amount)
	if err != nil {
		return primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
	}
	toAccount.Free = toFree

	if toAccount.Total().Lt(m.constants.ExistentialDeposit) {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorExistentialDeposit),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if err := m.ensureCanWithdraw(from, amount, primitives.ReasonsAll, fromAccount.Free); err != nil {
		return err
	}

	canDecProviders, err := m.Config.StoredMap.CanDecProviders(from)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	allowDeath := preservation == balancestypes.PreservationExpendable
	allowDeath = allowDeath && canDecProviders

	if !(allowDeath || fromAccount.Total().Gt(m.constants.ExistentialDeposit)) {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorKeepAlive),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	return nil
}
