package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callForceTransfer struct {
	primitives.Callable
	module module
}

func newCallForceTransfer(moduleId sc.U8, functionId sc.U8, module module) primitives.Call {
	return callForceTransfer{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}
}

func (c callForceTransfer) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	from, err := primitives.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}

	dest, err := primitives.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	value, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(
		from,
		dest,
		value,
	)

	return c, nil
}

func (c callForceTransfer) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callForceTransfer) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callForceTransfer) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callForceTransfer) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callForceTransfer) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callForceTransfer) BaseWeight() primitives.Weight {
	return callForceTransferWeight(c.module.constants.DbWeight)
}

func (_ callForceTransfer) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callForceTransfer) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callForceTransfer) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callForceTransfer) Docs() string {
	return "Exactly as `transfer_allow_death`, except the origin must be root and the source account may be specified."
}

func (c callForceTransfer) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsRootOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	fromMultiAddress, ok := args[0].(primitives.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid from value in callForceTransfer")
	}

	from, err := primitives.Lookup(fromMultiAddress)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	toMultiAddress, ok := args[1].(primitives.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid to value in callForceTransfer")
	}

	to, err := primitives.Lookup(toMultiAddress)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	valueCompact, ok := args[2].(sc.Compact)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid compact value in callForceTransfer")
	}
	value, ok := valueCompact.Number.(sc.U128)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid u128 value in callForceTransfer")
	}

	return primitives.PostDispatchInfo{}, c.module.transfer(from, to, value, types.PreservationExpendable)
}
