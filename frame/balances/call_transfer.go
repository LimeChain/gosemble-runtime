package balances

import (
	"bytes"
	"errors"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callTransfer struct {
	primitives.Callable
	transfer
}

func newCallTransfer(moduleId sc.U8, functionId sc.U8, storedMap primitives.StoredMap, constants *consts,
	mutator accountMutator) primitives.Call {
	call := callTransfer{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		transfer: newTransfer(moduleId, storedMap, constants, mutator),
	}

	return call
}

func (c callTransfer) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	dest, err := types.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	balance, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(
		dest,
		balance,
	)
	return c, nil
}

func (c callTransfer) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callTransfer) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callTransfer) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callTransfer) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callTransfer) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callTransfer) BaseWeight() types.Weight {
	return callTransferWeight(c.constants.DbWeight)
}

func (_ callTransfer) WeighData(baseWeight types.Weight) types.Weight {
	return types.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callTransfer) ClassifyDispatch(baseWeight types.Weight) types.DispatchClass {
	return types.NewDispatchClassNormal()
}

func (_ callTransfer) PaysFee(baseWeight types.Weight) types.Pays {
	return types.PaysYes
}

func (c callTransfer) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
	valueCompact, ok := args[1].(sc.Compact)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid compact value when dispatching call transfer")
	}
	value, ok := valueCompact.Number.(sc.U128)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid compact number field when dispatching call transfer")
	}
	return types.PostDispatchInfo{}, c.transfer.transfer(origin, args[0].(types.MultiAddress), value)
}

func (_ transfer) Docs() string {
	return "Transfer some liquid free balance to another account."
}

type transfer struct {
	moduleId       sc.U8
	storedMap      primitives.StoredMap
	constants      *consts
	accountMutator accountMutator
}

func newTransfer(moduleId sc.U8, storedMap primitives.StoredMap, constants *consts, mutator accountMutator) transfer {
	return transfer{
		moduleId:       moduleId,
		storedMap:      storedMap,
		constants:      constants,
		accountMutator: mutator,
	}
}

// transfer transfers liquid free balance from `source` to `dest`.
// Increases the free balance of `dest` and decreases the free balance of `origin` transactor.
// Must be signed by the transactor.
func (t transfer) transfer(origin types.RawOrigin, dest types.MultiAddress, value sc.U128) error {
	if !origin.IsSignedOrigin() {
		return types.NewDispatchErrorBadOrigin()
	}

	to, err := types.Lookup(dest)
	if err != nil {
		return types.NewDispatchErrorCannotLookup()
	}

	transactor, originErr := origin.AsSigned()
	if originErr != nil {
		return primitives.NewDispatchErrorOther(sc.Str(originErr.Error()))
	}

	return t.trans(transactor, to, value, types.ExistenceRequirementAllowDeath)
}

// trans transfers `value` free balance from `from` to `to`.
// Does not do anything if value is 0 or `from` and `to` are the same.
func (t transfer) trans(from types.AccountId, to types.AccountId, value sc.U128, existenceRequirement types.ExistenceRequirement) error {
	if value.Eq(constants.Zero) || reflect.DeepEqual(from, to) {
		return nil
	}

	_, err := t.accountMutator.tryMutateAccountWithDust(to, func(toAccount *types.AccountData, _ bool) (sc.Encodable, error) {
		return t.accountMutator.tryMutateAccountWithDust(from, func(fromAccount *types.AccountData, _ bool) (sc.Encodable, error) {
			return t.sanityChecks(from, fromAccount, toAccount, value, existenceRequirement)
		})
	})
	if err != nil {
		return err
	}

	t.storedMap.DepositEvent(newEventTransfer(t.moduleId, from, to, value))
	return nil
}

// sanityChecks checks the following:
// `fromAccount` has sufficient balance
// `toAccount` balance does not overflow
// `toAccount` total balance is more than the existential deposit
// `fromAccount` can withdraw `value`
// the existence requirements for `fromAccount`
// Updates the balances of `fromAccount` and `toAccount`.
func (t transfer) sanityChecks(from types.AccountId, fromAccount *types.AccountData, toAccount *types.AccountData, value sc.U128, existenceRequirement primitives.ExistenceRequirement) (sc.Encodable, error) {
	fromFree, err := sc.CheckedSubU128(fromAccount.Free, value)
	if err != nil {
		return nil, types.NewDispatchErrorModule(types.CustomModuleError{
			Index:   t.moduleId,
			Err:     sc.U32(ErrorInsufficientBalance),
			Message: sc.NewOption[sc.Str](nil),
		})
	}
	fromAccount.Free = fromFree

	toFree, err := sc.CheckedAddU128(toAccount.Free, value)
	if err != nil {
		return nil, types.NewDispatchErrorArithmetic(types.NewArithmeticErrorOverflow())
	}
	toAccount.Free = toFree

	if toAccount.Total().Lt(t.constants.ExistentialDeposit) {
		return nil, types.NewDispatchErrorModule(types.CustomModuleError{
			Index:   t.moduleId,
			Err:     sc.U32(ErrorExistentialDeposit),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if err := t.accountMutator.ensureCanWithdraw(from, value, types.ReasonsAll, fromAccount.Free); err != nil {
		return nil, err
	}

	canDecProviders, err := t.storedMap.CanDecProviders(from)
	if err != nil {
		return nil, types.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	allowDeath := existenceRequirement == types.ExistenceRequirementAllowDeath
	allowDeath = allowDeath && canDecProviders

	if !(allowDeath || fromAccount.Total().Gt(t.constants.ExistentialDeposit)) {
		return nil, types.NewDispatchErrorModule(types.CustomModuleError{
			Index:   t.moduleId,
			Err:     sc.U32(ErrorKeepAlive),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	return nil, nil
}

func (t transfer) reducibleBalance(who types.AccountId, keepAlive bool) (types.Balance, error) {
	account, err := t.storedMap.Get(who)
	if err != nil {
		return types.Balance{}, err
	}
	accountData := account.Data

	liquid := sc.SaturatingSubU128(accountData.Free, sc.Max128(accountData.FeeFrozen, accountData.MiscFrozen))
	canDecProviders, err := t.storedMap.CanDecProviders(who)
	if err != nil {
		return types.Balance{}, err
	}
	if canDecProviders && !keepAlive {
		return liquid, nil
	}

	mustRemainToExist := sc.SaturatingSubU128(t.constants.ExistentialDeposit, accountData.Total().Sub(liquid))
	return sc.SaturatingSubU128(liquid, mustRemainToExist), nil
}
