package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callForceSetBalance struct {
	primitives.Callable
	module Module
}

func newCallForceSetBalance(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	return callForceSetBalance{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}
}

func (c callForceSetBalance) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	who, err := primitives.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	value, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(
		who,
		value,
	)

	return c, nil
}

func (c callForceSetBalance) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callForceSetBalance) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callForceSetBalance) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callForceSetBalance) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callForceSetBalance) BaseWeight() primitives.Weight {
	return callForceSetBalanceCreatingWeight(c.module.constants.DbWeight).Max(callForceSetBalanceKillingWeight(c.module.constants.DbWeight))
}

func (_ callForceSetBalance) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callForceSetBalance) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callForceSetBalance) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callForceSetBalance) Docs() string {
	return "Set the regular balance of a given account. The dispatch origin for this call is `root`."
}

func (c callForceSetBalance) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsRootOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	dest, ok := args[0].(primitives.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid destination value in callForceSetBalance")
	}

	who, err := primitives.Lookup(dest)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	valueCompact, ok := args[1].(sc.Compact)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid compact value in callForceSetBalance")
	}
	value, ok := valueCompact.Number.(sc.U128)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid U128 value in callForceSetBalance")
	}

	return primitives.PostDispatchInfo{}, c.setBalance(who, value)
}

func (c callForceSetBalance) setBalance(who primitives.AccountId, newFree sc.U128) error {
	wipeOut := newFree.Lt(c.module.Config.ExistentialDeposit)
	if wipeOut {
		newFree = constants.Zero
	}

	result, err := c.module.mutateAccountHandlingDust(who, func(accountData *primitives.AccountData, bool bool) (sc.Encodable, error) {
		oldFree := accountData.Free
		accountData.Free = newFree

		return oldFree, nil
	})

	if err != nil {
		return err
	}
	oldFree, ok := result.(primitives.Balance)
	if !ok {
		return primitives.NewDispatchErrorOther("could not cast oldFree in callForceSetBalance")
	}

	if newFree.Gt(oldFree) {
		if err := newPositiveImbalance(newFree.Sub(oldFree), c.module.storage.TotalIssuance).Drop(); err != nil {
			return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		}
	} else if newFree.Lt(oldFree) {
		if err := newNegativeImbalance(oldFree.Sub(newFree), c.module.storage.TotalIssuance).Drop(); err != nil {
			return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		}
	}

	c.module.Config.StoredMap.DepositEvent(newEventBalanceSet(c.ModuleId, who, newFree))

	return nil
}
