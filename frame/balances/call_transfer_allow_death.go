package balances

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Transfer some liquid free balance to another account.
//
// `transfer_allow_death` will set the `FreeBalance` of the sender and receiver.
// If the sender's account is below the existential deposit as a result
// of the transfer, the account will be reaped.
//
// The dispatch origin for this call must be `Signed` by the transactor.
type callTransferAllowDeath struct {
	primitives.Callable
	module Module
}

func newCallTransferAllowDeath(functionId sc.U8, module Module) callTransferAllowDeath {
	call := callTransferAllowDeath{
		Callable: primitives.Callable{
			ModuleId:   module.Index,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}

	return call
}
func (c callTransferAllowDeath) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callTransferAllowDeath) Bytes() []byte {
	return c.Callable.Bytes()
}
func (c callTransferAllowDeath) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	dest, err := types.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	balance, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(
		dest,
		balance,
	)
	return c, nil
}

func (c callTransferAllowDeath) BaseWeight() types.Weight {
	return callTransferAllowDeathWeight(c.module.dbWeight())
}

func (_ callTransferAllowDeath) WeighData(baseWeight types.Weight) types.Weight {
	return types.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callTransferAllowDeath) ClassifyDispatch(baseWeight types.Weight) types.DispatchClass {
	return types.NewDispatchClassNormal()
}

func (_ callTransferAllowDeath) PaysFee(baseWeight types.Weight) types.Pays {
	return types.PaysYes
}

func (c callTransferAllowDeath) Docs() string {
	return "Transfer some liquid free balance to another account."
}

func (c callTransferAllowDeath) Dispatch(origin types.RuntimeOrigin, args sc.VaryingData) (types.PostDispatchInfo, error) {
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
		return types.PostDispatchInfo{}, errors.New("invalid compact value when dispatching call transfer_allow_death")
	}
	value, ok := valueCompact.Number.(sc.U128)
	if !ok {
		return types.PostDispatchInfo{}, errors.New("invalid compact number field when dispatching call transfer_allow_death")
	}

	if err := c.module.transfer(from, to, value, balancestypes.PreservationExpendable); err != nil {
		return types.PostDispatchInfo{}, err
	}

	return types.PostDispatchInfo{}, nil
}
