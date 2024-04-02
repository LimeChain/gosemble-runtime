package balances

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callForceUnreserve struct {
	primitives.Callable
	module Module
}

// / Unreserve some balance from a user by force.
// /
// / Can only be called by ROOT.
func newCallForceUnreserve(functionId sc.U8, module Module) callForceUnreserve {
	call := callForceUnreserve{
		Callable: primitives.Callable{
			ModuleId:   module.Index,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(types.MultiAddress{}, sc.U128{}),
		},
		module: module,
	}

	return call
}

func (c callForceUnreserve) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
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

func (c callForceUnreserve) BaseWeight() types.Weight {
	return callForceUnreserveWeight(c.module.Config.DbWeight)
}

func (_ callForceUnreserve) WeighData(baseWeight types.Weight) types.Weight {
	return types.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callForceUnreserve) ClassifyDispatch(baseWeight types.Weight) types.DispatchClass {
	return types.NewDispatchClassNormal()
}

func (_ callForceUnreserve) PaysFee(baseWeight types.Weight) types.Pays {
	return types.PaysYes
}

func (_ callForceUnreserve) Docs() string {
	return "Unreserve some balance from a user by force."
}

func (c callForceUnreserve) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
	if !origin.IsRootOrigin() {
		return types.PostDispatchInfo{}, types.NewDispatchErrorBadOrigin()
	}

	who, ok := args[0].(types.MultiAddress)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid who value when dispatching call force unreserve")
	}

	target, err := types.Lookup(who)
	if err != nil {
		c.module.logger.Debugf("Failed to lookup [%s]", who.Bytes())
		return types.PostDispatchInfo{}, types.NewDispatchErrorCannotLookup()
	}

	amount, ok := args[1].(sc.U128)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid amount value when dispatching call force unreserve")
	}

	if _, err := c.unreserve(target, amount); err != nil {
		return types.PostDispatchInfo{}, types.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	return types.PostDispatchInfo{}, nil
}

// / Unreserve some funds, returning any amount that was unable to be unreserved.
// /
// / Is a no-op if the value to be unreserved is zero or the account does not exist.
// /
// / NOTE: returns amount value which wasn't successfully unreserved.
// forceUnreserve frees funds, returning the amount that has not been freed.
func (c callForceUnreserve) unreserve(who primitives.AccountId, value primitives.Balance) (primitives.Balance, error) {
	if value.Eq(constants.Zero) {
		return constants.Zero, nil
	}

	account, err := c.module.Config.StoredMap.Get(who)
	if err != nil {
		return sc.U128{}, err
	}

	totalBalance := account.Data.Total()
	if totalBalance.Eq(constants.Zero) {
		return value, nil
	}

	actual := sc.Min128(account.Data.Reserved, value)
	account.Data.Reserved = account.Data.Reserved.Sub(actual)
	account.Data.Free = sc.SaturatingAddU128(account.Data.Free, actual)

	result, err := c.module.tryMutateAccountNew(who, account.Data)
	if err != nil {
		// This should never happen since we don't alter the total amount in the account.
		// If it ever does, then we should fail gracefully though, indicating that nothing
		// could be done.
		return sc.NewU128(0), err
	}

	c.module.Config.StoredMap.DepositEvent(newEventUnreserved(c.ModuleId, who, result))

	return value.Sub(actual), nil
}
