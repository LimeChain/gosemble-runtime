package babe

import (
	"bytes"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// The epoch config change is recorded and will be enacted on
// the next call to `enact_epoch_change`. The config will be activated one epoch after.
// Multiple calls to this method will replace any existing planned config change that had
// not been enacted yet.
type callPlanConfigChange struct {
	primitives.Callable
	storagePendingEpochConfigChange support.StorageValue[NextConfigDescriptor]
}

func newCallPlanConfigChange(moduleId sc.U8, functionId sc.U8, storagePendingEpochConfigChange support.StorageValue[NextConfigDescriptor]) primitives.Call {
	call := callPlanConfigChange{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(NextConfigDescriptor{}),
		},
		storagePendingEpochConfigChange: storagePendingEpochConfigChange,
	}

	return call
}

func (c callPlanConfigChange) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	nextConfigDescriptor, err := DecodeNextConfigDescriptor(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(nextConfigDescriptor)
	return c, nil
}

func (c callPlanConfigChange) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callPlanConfigChange) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callPlanConfigChange) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callPlanConfigChange) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callPlanConfigChange) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callPlanConfigChange) BaseWeight() primitives.Weight {
	return callPlanConfigChangeWeight(primitives.RuntimeDbWeight{})
}

func (_ callPlanConfigChange) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callPlanConfigChange) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callPlanConfigChange) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callPlanConfigChange) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	// TODO: enable once 'sudo' module is implemented
	//
	// err := EnsureRoot(origin)
	// if err != nil {
	// 	return primitives.PostDispatchInfo{}, err
	// }

	config := args[0].(NextConfigDescriptor)

	if reflect.TypeOf(config) == reflect.TypeOf(NextConfigDescriptor{}) && reflect.TypeOf(config.V1) == reflect.TypeOf(babetypes.EpochConfiguration{}) {
		if !((config.V1.C.Numerator != 0 || !reflect.DeepEqual(config.V1.AllowedSlots, babetypes.NewPrimarySlots())) && config.V1.C.Denominator != 0) {
			return primitives.PostDispatchInfo{}, NewDispatchErrorInvalidConfiguration(c.ModuleId)
		}
	}

	c.storagePendingEpochConfigChange.Put(config)

	return primitives.PostDispatchInfo{}, nil
}

func (_ callPlanConfigChange) Docs() string {
	return "Plan an epoch config change."
}
