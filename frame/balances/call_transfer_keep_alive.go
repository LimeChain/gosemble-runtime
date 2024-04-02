package balances

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Same as the [`transfer_allow_death`] call, but with a check that the transfer will not
// kill the origin account.
//
// 99% of the time you want [`transfer_allow_death`] instead.
//
// [`transfer_allow_death`]: struct.Pallet.html#method.transfer
type callTransferKeepAlive struct {
	primitives.Callable
	module Module
}

func newCallTransferKeepAlive(functionId sc.U8, module Module) callTransferKeepAlive {
	call := callTransferKeepAlive{
		Callable: primitives.Callable{
			ModuleId:   module.Index,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(types.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}

	return call
}

func (c callTransferKeepAlive) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	dest, err := types.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	value, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(
		dest,
		value,
	)
	return c, nil
}

func (c callTransferKeepAlive) BaseWeight() types.Weight {
	return callTransferKeepAliveWeight(c.module.constants.DbWeight)
}

func (_ callTransferKeepAlive) WeighData(baseWeight types.Weight) types.Weight {
	return types.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callTransferKeepAlive) ClassifyDispatch(baseWeight types.Weight) types.DispatchClass {
	return types.NewDispatchClassNormal()
}

func (_ callTransferKeepAlive) PaysFee(baseWeight types.Weight) types.Pays {
	return types.PaysYes
}

func (_ callTransferKeepAlive) Docs() string {
	return "Same as the [`transfer_allow_death`] call, but with a check that the transfer will not kill the origin account."
}

func (c callTransferKeepAlive) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	from, err := origin.AsSigned()
	if err != nil {
		c.module.logger.Warnf("err dispatch transfer_allow_death: %v", err)

		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	dest, ok := args[0].(types.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, types.NewDispatchErrorCannotLookup()
	}

	to, err := primitives.Lookup(dest)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}
	valueCompact, ok := args[1].(sc.Compact)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid compact value when dispatching call transfer_keep_alive")
	}
	value, ok := valueCompact.Number.(sc.U128)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid compact number field when dispatching call transfer_keep_alive")
	}

	if err := c.module.transfer(from, to, value, balancestypes.PreservationExpendable); err != nil {
		return types.PostDispatchInfo{}, err
	}

	return types.PostDispatchInfo{}, nil
}
