package balances

import (
	"bytes"
	"errors"
	"math"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	errInvalidAccountIdSequence = errors.New("invalid accountId sequence value when dispatching call upgrade accounts")
)

// Upgrade a specified account.
//
// - `origin`: Must be `Signed`.
// - `who`: The account to be upgraded.

// This will waive the transaction fee if at least all but 10% of the accounts needed to
// be upgraded. (We let some not have to be upgraded just in order to allow for the
// possibililty of churn).
type callUpgradeAccounts struct {
	primitives.Callable
	module Module
}

func newCallUpgradeAccounts(functionId sc.U8, module Module) callUpgradeAccounts {
	call := callUpgradeAccounts{
		Callable: primitives.Callable{
			ModuleId:   module.moduleIndex(),
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.Sequence[sc.U8]]{}),
		},
		module: module,
	}

	return call
}

func (c callUpgradeAccounts) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	who, err := sc.DecodeSequenceWith(buffer, sc.DecodeSequence[sc.U8])
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(
		who,
	)

	return c, nil
}

func (c callUpgradeAccounts) BaseWeight() primitives.Weight {
	who, ok := c.Arguments[0].(sc.Sequence[sc.Sequence[sc.U8]])
	size := 0
	if ok {
		size = len(who)
	}
	return callUpgradeAccountsWeight(c.module.dbWeight(), sc.U64(size))
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
	return "Upgrade a specified account."
}

func (c callUpgradeAccounts) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	who, ok := args[0].(sc.Sequence[sc.Sequence[sc.U8]])
	if !ok {
		return primitives.PostDispatchInfo{}, errInvalidAccountIdSequence
	}

	pays, err := c.upgradeAccounts(who)
	return primitives.PostDispatchInfo{PaysFee: pays}, err
}

func (c callUpgradeAccounts) upgradeAccounts(who sc.Sequence[sc.Sequence[sc.U8]]) (primitives.Pays, error) {
	if len(who) == 0 {
		return primitives.PaysYes, nil
	}

	upgradeCount := 0
	for _, accId := range who {
		accId, err := primitives.NewAccountId(accId...)
		if err != nil {
			return primitives.PaysNo, err
		}

		upgraded, err := c.module.ensureUpgraded(accId)
		if err != nil {
			return primitives.PaysNo, err
		}

		if upgraded {
			upgradeCount++
		}
	}

	if float64(upgradeCount) >= math.Floor(float64(len(who))*0.9) {
		return primitives.PaysNo, nil
	} else {
		return primitives.PaysYes, nil
	}
}
