package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callTransferAllowDeath struct {
	primitives.Callable
	module Module
}

func newCallTransferAllowDeath(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	return callTransferAllowDeath{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}
}

func (c callTransferAllowDeath) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
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

func (c callTransferAllowDeath) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callTransferAllowDeath) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callTransferAllowDeath) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callTransferAllowDeath) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callTransferAllowDeath) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callTransferAllowDeath) BaseWeight() primitives.Weight {
	return callTransferAllowDeathWeight(c.module.constants.DbWeight)
}

func (_ callTransferAllowDeath) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callTransferAllowDeath) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callTransferAllowDeath) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callTransferAllowDeath) Docs() string {
	return "Transfer some liquid free balance to another account. " +
		"`transfer_allow_death` will set the `FreeBalance` of the sender and receiver. " +
		" If the sender's account is below the existential deposit as a result" +
		" of the transfer, the account will be reaped." +
		"The dispatch origin for this call must be `Signed` by the transactor."
}

func (c callTransferAllowDeath) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	from, originErr := origin.AsSigned()
	if originErr != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(originErr.Error()))
	}

	dest, ok := args[0].(primitives.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid destination value in callTransferAllowDeath")
	}

	to, err := primitives.Lookup(dest)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	valueCompact, ok := args[1].(sc.Compact)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid compact value in callTransferAllowDeath")
	}
	value, ok := valueCompact.Number.(sc.U128)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid U128 value in callTransferAllowDeath")
	}

	return primitives.PostDispatchInfo{}, c.module.transfer(from, to, value, types.PreservationExpendable)
}
