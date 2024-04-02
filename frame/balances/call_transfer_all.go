package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Transfer the entire transferable balance from the caller account.
//
// NOTE: This function only attempts to transfer _transferable_ balances. This means that
// any locked, reserved, or existential deposits (when `keep_alive` is `true`), will not be
// transferred by this function. To ensure that this function results in a killed account,
// you might need to prepare the account by removing any reference counters, storage
// deposits, etc...
//
// The dispatch origin of this call must be Signed.
//
//   - `dest`: The recipient of the transfer.
//   - `keep_alive`: A boolean to determine if the `transfer_all` operation should send all
//     of the funds the account has, causing the sender account to be killed (false), or
//     transfer everything except at least the existential deposit, which will guarantee to
//     keep the sender account alive (true).
type callTransferAll struct {
	primitives.Callable
	module Module
}

func newCallTransferAll(functionId sc.U8, module Module) callTransferAll {
	call := callTransferAll{
		Callable: primitives.Callable{
			ModuleId:   module.Index,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(types.MultiAddress{}, sc.Bool(true)),
		},
		module: module,
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

func (c callTransferAll) BaseWeight() types.Weight {
	return callTransferAllWeight(c.module.dbWeight())
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

func (_ callTransferAll) Docs() string {
	return "Transfer the entire transferable balance from the caller account."
}

// transferAll transfers the entire transferable balance from `origin` to `dest`.
// By transferable it means that any locked or reserved amounts will not be transferred.
// `keepAlive`: A boolean to determine if the `transfer_all` operation should send all
// the funds the account has, causing the sender account to be killed (false), or
// transfer everything except at least the existential deposit, which will guarantee to
// keep the sender account alive (true).
func (c callTransferAll) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	from, err := origin.AsSigned()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	dest, ok := args[0].(types.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, types.NewDispatchErrorCannotLookup()
	}

	keepAliveArg, ok := args[1].(sc.Bool)
	if !ok {
		return primitives.PostDispatchInfo{}, types.NewDispatchErrorOther(sc.Str("Failed to decode keepAlive"))
	}

	keepAlive := balancestypes.PreservationExpendable
	if keepAliveArg {
		keepAlive = balancestypes.PreservationPreserve
	}

	reducibleBalance, err := c.module.reducibleBalance(from, keepAlive, false)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	to, err := types.Lookup(dest)
	if err != nil {
		return primitives.PostDispatchInfo{}, types.NewDispatchErrorCannotLookup()
	}

	if err := c.module.transfer(from, to, reducibleBalance, keepAlive); err != nil {
		return types.PostDispatchInfo{}, err
	}

	return types.PostDispatchInfo{}, nil
}
