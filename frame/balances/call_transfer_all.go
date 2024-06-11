package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callTransferAll struct {
	primitives.Callable
	transfer
	logger log.RuntimeLogger
}

func newCallTransferAll(moduleId sc.U8, functionId sc.U8, storedMap primitives.StoredMap, constants *consts, mutator accountMutator, logger log.RuntimeLogger) primitives.Call {
	call := callTransferAll{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(types.MultiAddress{}, sc.Bool(true)),
		},
		transfer: newTransfer(moduleId, storedMap, constants, mutator),
		logger:   logger,
	}

	return call
}

func (c callTransferAll) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	dest, err := types.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	keepAlive, err := sc.DecodeBool(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(
		dest,
		keepAlive,
	)
	return c, nil
}

func (c callTransferAll) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callTransferAll) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callTransferAll) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callTransferAll) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callTransferAll) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callTransferAll) BaseWeight() types.Weight {
	return callTransferAllWeight(c.constants.DbWeight)
}

func (_ callTransferAll) WeighData(baseWeight types.Weight) types.Weight {
	return types.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callTransferAll) ClassifyDispatch(baseWeight types.Weight) types.DispatchClass {
	return types.NewDispatchClassNormal()
}

func (_ callTransferAll) PaysFee(baseWeight types.Weight) types.Pays {
	return types.PaysYes
}

func (c callTransferAll) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
	return types.PostDispatchInfo{}, c.transferAll(origin, args[0].(types.MultiAddress), bool(args[1].(sc.Bool)))
}

func (_ callTransferAll) Docs() string {
	return "Transfer the entire transferable balance from the caller account."
}

// transferAll transfers the entire transferable balance from `origin` to `dest`.
// By transferable it means that any locked or reserved amounts will not be transferred.
// `keepAlive`: A boolean to determine if the `transfer_all` operation should send all
// the funds the account has, causing the sender account to be killed (false), or
// transfer everything except at least the existential deposit, which will guarantee to
// keep the sender account alive (true).
func (c callTransferAll) transferAll(origin types.RawOrigin, dest types.MultiAddress, keepAlive bool) error {
	if !origin.IsSignedOrigin() {
		return types.NewDispatchErrorBadOrigin()
	}

	transactor, err := origin.AsSigned()
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	reducibleBalance, err := c.reducibleBalance(transactor, keepAlive)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	to, errLookup := types.Lookup(dest)
	if errLookup != nil {
		c.logger.Debugf("Failed to lookup [%s]", dest.Bytes())
		return types.NewDispatchErrorCannotLookup()
	}

	keep := types.ExistenceRequirementKeepAlive
	if !keepAlive {
		keep = types.ExistenceRequirementAllowDeath
	}

	return c.transfer.trans(transactor, to, reducibleBalance, keep)
}
