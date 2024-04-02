package balances

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

// Set the regular balance of a given account.
//
// The dispatch origin for this call is `root`.
type callForceSetBalance struct {
	types.Callable
	module Module
}

func newCallForceSetBalance(functionId sc.U8, module Module) callForceSetBalance {
	call := callForceSetBalance{
		Callable: types.Callable{
			ModuleId:   module.Index,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(types.MultiAddress{}, sc.Compact{Number: sc.U128{}}, sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}

	return call
}

func (c callForceSetBalance) DecodeArgs(buffer *bytes.Buffer) (types.Call, error) {
	targetAddress, err := types.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	newFree, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}
	newReserved, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(
		targetAddress,
		newFree,
		newReserved,
	)
	return c, nil
}

func (c callForceSetBalance) BaseWeight() types.Weight {
	return callForceSetBalanceCreatingWeight(c.module.constants.DbWeight).Max(callForceSetBalanceKillingWeight(c.module.constants.DbWeight))
}

func (_ callForceSetBalance) IsInherent() bool {
	return false
}

func (_ callForceSetBalance) WeighData(baseWeight types.Weight) types.Weight {
	return types.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callForceSetBalance) ClassifyDispatch(baseWeight types.Weight) types.DispatchClass {
	return types.NewDispatchClassNormal()
}

func (_ callForceSetBalance) PaysFee(baseWeight types.Weight) types.Pays {
	return types.PaysYes
}

func (_ callForceSetBalance) Docs() string {
	return "Set the balances of a given account."
}

func (c callForceSetBalance) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
	compactFree, ok := args[1].(sc.Compact)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid free compact value when dispatching balance call set")
	}
	newFree, ok := compactFree.Number.(sc.U128)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid free compact number when dispatching balance call set")
	}

	compactReserved, ok := args[2].(sc.Compact)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid reserved compact value when dispatching balance call set")
	}
	newReserved, ok := compactReserved.Number.(sc.U128)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid reserved compact number when dispatching balance call set")
	}
	return types.PostDispatchInfo{}, c.setBalance(origin, args[0].(types.MultiAddress), newFree, newReserved)
}

// setBalance sets the balance of a given account.
// Changes free and reserve balance of `who`,
// including the total issuance.
// Can only be called by ROOT.
func (c callForceSetBalance) setBalance(origin types.RawOrigin, who types.MultiAddress, newFree sc.U128, newReserved sc.U128) error {
	if !origin.IsRootOrigin() {
		return types.NewDispatchErrorBadOrigin()
	}

	address, err := types.Lookup(who)
	if err != nil {
		return types.NewDispatchErrorCannotLookup()
	}

	sum := newFree.Add(newReserved)

	if sum.Lt(c.module.constants.ExistentialDeposit) {
		newFree = sc.NewU128(0)
		newReserved = sc.NewU128(0)
	}

	// new code
	account, err := c.module.Config.StoredMap.Get(address)
	if err != nil {
		return err
	}
	oldFree := account.Data.Free
	oldReserved := account.Data.Reserved

	account.Data.Free = newFree
	account.Data.Reserved = newReserved
	// old code
	_, err = c.module.tryMutateAccountNew(address, account.Data)
	if err != nil {
		return err
	}

	// parsedResult := result.(sc.VaryingData)
	// oldFree := parsedResult[0].(types.Balance)
	// oldReserved := parsedResult[1].(types.Balance)

	if newFree.Gt(oldFree) {
		if err := newPositiveImbalance(newFree.Sub(oldFree), c.module.storage.TotalIssuance).Drop(); err != nil {
			return types.NewDispatchErrorOther(sc.Str(err.Error()))
		}

	} else if newFree.Lt(oldFree) {
		if err := newNegativeImbalance(oldFree.Sub(newFree), c.module.storage.TotalIssuance).Drop(); err != nil {
			return types.NewDispatchErrorOther(sc.Str(err.Error()))
		}
	}

	if newReserved.Gt(oldReserved) {
		if err := newPositiveImbalance(newReserved.Sub(oldReserved), c.module.storage.TotalIssuance).Drop(); err != nil {
			return types.NewDispatchErrorOther(sc.Str(err.Error()))
		}
	} else if newReserved.Lt(oldReserved) {
		if err := newNegativeImbalance(oldReserved.Sub(newReserved), c.module.storage.TotalIssuance).Drop(); err != nil {
			return types.NewDispatchErrorOther(sc.Str(err.Error()))
		}

	}

	whoAccountId, errAccId := who.AsAccountId()
	if errAccId != nil {
		return types.NewDispatchErrorOther(sc.Str(errAccId.Error()))
	}

	c.module.Config.StoredMap.DepositEvent(
		newEventBalanceSet(
			c.ModuleId,
			whoAccountId,
			newFree,
			newReserved,
		),
	)
	return nil
}

// updateAccount updates the reserved and free amounts and returns the old amounts
func updateAccount(account *types.AccountData, newFree, newReserved sc.U128) (oldFree, oldReserved types.Balance) {
	oldFree = account.Free
	oldReserved = account.Reserved

	account.Free = newFree
	account.Reserved = newReserved

	return oldFree, oldReserved
}
