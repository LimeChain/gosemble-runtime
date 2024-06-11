package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callTransferKeepAlive struct {
	primitives.Callable
	module Module
}

func newCallTransferKeepAlive(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	return callTransferKeepAlive{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}
}

func (c callTransferKeepAlive) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	dest, err := primitives.DecodeMultiAddress(buffer)
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

func (c callTransferKeepAlive) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callTransferKeepAlive) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callTransferKeepAlive) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callTransferKeepAlive) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callTransferKeepAlive) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callTransferKeepAlive) BaseWeight() primitives.Weight {
	return callTransferKeepAliveWeight(c.module.constants.DbWeight)
}

func (_ callTransferKeepAlive) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callTransferKeepAlive) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callTransferKeepAlive) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callTransferKeepAlive) Docs() string {
	return "Same as the [`transfer_allow_death`] call, but with a check that the transfer will not " +
		"kill the origin account. " +
		"99% of the time you want [`transfer_allow_death`] instead. " +
		"[`transfer_allow_death`]: struct.Pallet.html#method.transfer"
}

func (c callTransferKeepAlive) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	from, originErr := origin.AsSigned()
	if originErr != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(originErr.Error()))
	}

	dest, ok := args[0].(primitives.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid destination value in callTransferKeepAlive")
	}

	to, err := primitives.Lookup(dest)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	valueCompact, ok := args[1].(sc.Compact)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid compact value in callTransferKeepAlive")
	}
	value, ok := valueCompact.Number.(sc.U128)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid U128 value in callTransferKeepAlive")
	}

	return primitives.PostDispatchInfo{}, c.module.transfer(from, to, value, types.PreservationPreserve)
}
