package balances

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	// "github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	errInvalidArgAdjustmentDirection = errors.New("invalid adjustment direction arg when dispatching call force adjust total issuance")
	errInvalidArgDeltaCompact        = errors.New("invalid compact value when dispatching call force adjust total issuance")
	errInvalidArgDelta               = errors.New("invalid compact number field when dispatching call force adjust total issuance")
)

// Adjust the total issuance in a saturating way.
//
// Can only be called by root and always needs a positive `delta`.
type callForceAdjustTotalIssuance struct {
	primitives.Callable
	module Module
}

func newCallForceAdjustTotalIssuance(functionId sc.U8, module Module) callForceAdjustTotalIssuance {
	call := callForceAdjustTotalIssuance{
		Callable: primitives.Callable{
			ModuleId:   module.moduleIndex(),
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(balancestypes.AdjustmentDirection(0), sc.Compact{Number: sc.U128{}}),
		},
		module: module,
	}

	return call
}

func (c callForceAdjustTotalIssuance) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	adjustmentDirection, err := balancestypes.DecodeAdjustmentDirection(buffer)
	if err != nil {
		return nil, err
	}
	delta, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(
		adjustmentDirection,
		delta,
	)

	return c, nil
}

func (c callForceAdjustTotalIssuance) BaseWeight() primitives.Weight {
	return callForceAdjustTotalIssuanceWeight(c.module.dbWeight())
}

func (_ callForceAdjustTotalIssuance) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callForceAdjustTotalIssuance) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callForceAdjustTotalIssuance) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callForceAdjustTotalIssuance) Docs() string {
	return "Adjust the total issuance in a saturating way."
}

func (c callForceAdjustTotalIssuance) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsRootOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	adjustmentDirection, ok := args[0].(balancestypes.AdjustmentDirection)
	if !ok {
		return primitives.PostDispatchInfo{}, errInvalidArgAdjustmentDirection
	}

	deltaCompact, ok := args[1].(sc.Compact)
	if !ok {
		return primitives.PostDispatchInfo{}, errInvalidArgDeltaCompact
	}
	delta, ok := deltaCompact.Number.(sc.U128)
	if !ok {
		return primitives.PostDispatchInfo{}, errInvalidArgDelta
	}

	if err := c.forceAdjustTotalIssuance(delta, adjustmentDirection); err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	return primitives.PostDispatchInfo{}, nil
}

func (c callForceAdjustTotalIssuance) forceAdjustTotalIssuance(delta sc.U128, direction balancestypes.AdjustmentDirection) error {
	if !delta.Gt(constants.Zero) {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   c.ModuleId,
			Err:     sc.U32(ErrorDeltaZero),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	totalIssuance, err := c.module.storage.TotalIssuance.Get()
	if err != nil {
		return err
	}

	var newTotalIssuance sc.U128
	if direction == balancestypes.AdjustmentDirectionIncrease {
		newTotalIssuance = sc.SaturatingAddU128(totalIssuance, delta)
	} else {
		newTotalIssuance = sc.SaturatingSubU128(totalIssuance, delta)
	}

	inactiveIssuance, err := c.module.storage.InactiveIssuance.Get()
	if err != nil {
		return err
	}
	if inactiveIssuance.Gt(newTotalIssuance) {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   c.module.Index,
			Err:     sc.U32(IssuanceDeactivated),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	c.module.storage.TotalIssuance.Put(newTotalIssuance)

	c.module.Config.StoredMap.DepositEvent(newEventTotalIssuanceForced(c.module.Index, totalIssuance, newTotalIssuance))

	return nil
}
