package balances

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Exactly as `transfer_allow_death`, except the origin must be root and the source account
// may be specified.
type callForceTransfer struct {
	primitives.Callable
	module Module
}

func newCallForceTransfer(functionId sc.U8, module Module) callForceTransfer {
	call := callForceTransfer{
		Callable: primitives.Callable{
			ModuleId:   module.Index,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(types.MultiAddress{}, types.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}

	return call
}

func (c callForceTransfer) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	source, err := types.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	dest, err := types.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	value, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(
		source,
		dest,
		value,
	)
	return c, nil
}

func (c callForceTransfer) BaseWeight() types.Weight {
	return callForceTransferWeight(c.module.dbWeight())
}

func (_ callForceTransfer) WeighData(baseWeight types.Weight) types.Weight {
	return types.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callForceTransfer) ClassifyDispatch(baseWeight types.Weight) types.DispatchClass {
	return types.NewDispatchClassNormal()
}

func (_ callForceTransfer) PaysFee(baseWeight types.Weight) types.Pays {
	return types.PaysYes
}

func (_ callForceTransfer) Docs() string {
	return "Exactly as `transfer_allow_death`, except the origin must be root and the source account may be specified."
}

// forceTransfer transfers liquid free balance from `source` to `dest`.
// Can only be called by ROOT.
func (c callForceTransfer) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
	if !origin.IsRootOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	source, ok := args[0].(types.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, types.NewDispatchErrorCannotLookup()
	}
	from, err := primitives.Lookup(source)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	dest, ok := args[1].(types.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, types.NewDispatchErrorCannotLookup()
	}
	to, err := primitives.Lookup(dest)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	valueCompact, ok := args[2].(sc.Compact)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid Compact value when dispatching call_force_transfer")
	}
	value, ok := valueCompact.Number.(sc.U128)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid Compact field number when dispatching call_force_transfer")
	}

	if err := c.module.transfer(from, to, value, balancestypes.PreservationExpendable); err != nil {
		return types.PostDispatchInfo{}, err
	}

	return types.PostDispatchInfo{}, nil
}
