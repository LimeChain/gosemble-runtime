package balances

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callForceFree struct {
	primitives.Callable
	transfer
	logger log.RuntimeLogger
}

func newCallForceFree(moduleId sc.U8, functionId sc.U8, storedMap primitives.StoredMap, constants *consts, mutator accountMutator, logger log.RuntimeLogger) primitives.Call {
	call := callForceFree{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(types.MultiAddress{}, sc.U128{}),
		},
		transfer: newTransfer(moduleId, storedMap, constants, mutator),
		logger:   logger,
	}

	return call
}

func (c callForceFree) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	who, err := types.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	amount, err := sc.DecodeU128(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(
		who,
		amount,
	)
	return c, nil
}

func (c callForceFree) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callForceFree) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callForceFree) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callForceFree) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callForceFree) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callForceFree) BaseWeight() types.Weight {
	return callForceFreeWeight(c.constants.DbWeight)
}

func (_ callForceFree) WeighData(baseWeight types.Weight) types.Weight {
	return types.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callForceFree) ClassifyDispatch(baseWeight types.Weight) types.DispatchClass {
	return types.NewDispatchClassNormal()
}

func (_ callForceFree) PaysFee(baseWeight types.Weight) types.Pays {
	return types.PaysYes
}

func (_ callForceFree) Docs() string {
	return "Unreserve some balance from a user by force."
}

func (c callForceFree) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
	amount, ok := args[1].(sc.U128)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid amount value when dispatching call force free")
	}
	return types.PostDispatchInfo{}, c.forceFree(origin, args[0].(types.MultiAddress), amount)
}

// forceFree frees some balance from a user by force.
// Can only be called by ROOT.
// Consider Substrate fn force_unreserve
func (c callForceFree) forceFree(origin types.RawOrigin, who types.MultiAddress, amount sc.U128) error {
	if !origin.IsRootOrigin() {
		return types.NewDispatchErrorBadOrigin()
	}

	target, err := types.Lookup(who)
	if err != nil {
		c.logger.Debugf("Failed to lookup [%s]", who.Bytes())
		return types.NewDispatchErrorCannotLookup()
	}

	if _, err := c.force(target, amount); err != nil {
		return types.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	return nil
}

// forceFree frees funds, returning the amount that has not been freed.
func (c callForceFree) force(who primitives.AccountId, value sc.U128) (sc.U128, error) {
	if value.Eq(constants.Zero) {
		return constants.Zero, nil
	}

	account, err := c.storedMap.Get(who)
	if err != nil {
		return sc.U128{}, err
	}

	totalBalance := account.Data.Total()
	if totalBalance.Eq(constants.Zero) {
		return value, nil
	}

	result, err := c.accountMutator.tryMutateAccount(
		who,
		func(account *types.AccountData, _ bool) (sc.Encodable, error) {
			return removeReserveAndFree(account, value), nil
		},
	)

	if err != nil {
		return sc.NewU128(0), err
	}

	actual := result.(primitives.Balance)
	c.storedMap.DepositEvent(newEventUnreserved(c.ModuleId, who, actual))

	return value.Sub(actual), nil
}

// removeReserveAndFree frees reserved value from the account.
func removeReserveAndFree(account *types.AccountData, value sc.U128) primitives.Balance {
	actual := sc.Min128(account.Reserved, value)
	account.Reserved = account.Reserved.Sub(actual)

	account.Free = sc.SaturatingAddU128(account.Free, actual)

	return actual
}
