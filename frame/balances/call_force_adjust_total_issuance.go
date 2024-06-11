package balances

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callForceAdjustTotalIssuance struct {
	primitives.Callable
	storage        *storage
	eventDepositor primitives.EventDepositor
}

func newCallForceAdjustTotalIssuance(moduleId sc.U8, functionId sc.U8, eventDepositor primitives.EventDepositor, storage *storage) primitives.Call {
	return callForceAdjustTotalIssuance{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(types.AdjustDirection{}, sc.Compact{Number: sc.U128{}}),
		},
		eventDepositor: eventDepositor,
		storage:        storage,
	}
}

func (c callForceAdjustTotalIssuance) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	direction, err := types.DecodeAdjustDirection(buffer)
	if err != nil {
		return nil, err
	}
	delta, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(
		direction,
		delta,
	)

	return c, nil
}

func (c callForceAdjustTotalIssuance) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callForceAdjustTotalIssuance) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callForceAdjustTotalIssuance) ModuleIndex() sc.U8 { return c.Callable.ModuleIndex() }

func (c callForceAdjustTotalIssuance) FunctionIndex() sc.U8 { return c.Callable.FunctionIndex() }

func (c callForceAdjustTotalIssuance) Args() sc.VaryingData { return c.Callable.Args() }

func (c callForceAdjustTotalIssuance) BaseWeight() primitives.Weight {
	return callForceAdjustTotalIssuanceWeight()
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
	return "Adjust the total issuance in a saturating way. Can only be called by root and always needs a positive `delta`."
}

func (c callForceAdjustTotalIssuance) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsRootOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}
	direction, ok := args[0].(types.AdjustDirection)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid direction value when dispatching callForceAdjustTotalIssuance")
	}
	deltaCompact, ok := args[1].(sc.Compact)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid amount value when dispatching call force free")
	}

	delta, ok := deltaCompact.Number.(sc.U128)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid Compact field delta when dispatch call_force_adjust_total_issuance")
	}

	if delta.Lte(constants.Zero) {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   c.ModuleId,
			Err:     sc.U32(ErrorDeltaZero),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	totalIssuance, err := c.storage.TotalIssuance.Get()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	var newIssuance sc.U128
	if direction.IsIncrease() {
		newIssuance = sc.SaturatingAddU128(totalIssuance, delta)
	} else {
		newIssuance = sc.SaturatingSubU128(totalIssuance, delta)
	}

	inactiveIssuance, err := c.storage.InactiveIssuance.Get()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	if inactiveIssuance.Gt(newIssuance) {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   c.ModuleId,
			Err:     sc.U32(ErrorIssuanceDeactivated),
			Message: sc.NewOption[sc.Str](nil),
		})
	}
	c.storage.TotalIssuance.Put(newIssuance)
	c.eventDepositor.DepositEvent(newEventTotalIssuanceForced(c.ModuleId, totalIssuance, newIssuance))

	return primitives.PostDispatchInfo{}, nil
}
