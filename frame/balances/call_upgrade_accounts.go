package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callUpgradeAccounts struct {
	primitives.Callable
	module Module
}

func newCallUpgradeAccounts(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	return callUpgradeAccounts{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[primitives.AccountId]{}),
		},
		module: module,
	}
}

func (c callUpgradeAccounts) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	who, err := primitives.DecodeSequenceAccountId(buffer)
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(who)

	return c, nil
}

func (c callUpgradeAccounts) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callUpgradeAccounts) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callUpgradeAccounts) ModuleIndex() sc.U8 { return c.Callable.ModuleIndex() }

func (c callUpgradeAccounts) FunctionIndex() sc.U8 { return c.Callable.FunctionIndex() }

func (c callUpgradeAccounts) Args() sc.VaryingData { return c.Callable.Args() }

func (c callUpgradeAccounts) BaseWeight() primitives.Weight {
	accounts := c.Arguments[0].(sc.Sequence[primitives.AccountId])
	return callUpgradeAccountsWeight(c.module.constants.DbWeight, sc.U64(len(accounts)))
}

func (_ callUpgradeAccounts) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callUpgradeAccounts) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callUpgradeAccounts) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callUpgradeAccounts) Docs() string {
	return "Upgrade a specified `account`."
}

func (c callUpgradeAccounts) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	who, ok := args[0].(sc.Sequence[primitives.AccountId])
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid argument in callUpgradeAccounts")
	}
	if len(who) == 0 {
		return primitives.PostDispatchInfo{PaysFee: primitives.PaysYes}, nil
	}

	upgradeCount := 0
	for _, accountId := range who {
		upgraded, err := c.module.ensureUpgraded(accountId)
		if err != nil {
			return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		}
		if upgraded {
			upgradeCount += 1
		}
	}
	upgraded := (upgradeCount * 100) / len(who)
	if upgraded > 90 {
		return primitives.PostDispatchInfo{PaysFee: primitives.PaysNo}, nil
	}

	return primitives.PostDispatchInfo{PaysFee: primitives.PaysYes}, nil
}
