package babe

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	nextConfig = NextConfigDescriptor{
		V1: babetypes.EpochConfiguration{
			C:            types.Tuple2U64{First: 1, Second: 2},
			AllowedSlots: babetypes.NewPrimarySlots(),
		},
	}

	somePlanConfigChangeArgs = sc.NewVaryingData(nextConfig)

	defaultPlanConfigChangeArgs = sc.NewVaryingData(NextConfigDescriptor{})
	baseWeight                  = primitives.WeightFromParts(567, 123)
)

func Test_Call_PlanConfigChange_New(t *testing.T) {
	call := setupCallPlanConfigChange()

	expected := callPlanConfigChange{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionPlanConfigChangeIndex,
			Arguments:  defaultPlanConfigChangeArgs,
		},
		dbWeight:                        dbWeight,
		storagePendingEpochConfigChange: mockStoragePendingEpochConfigChange,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_PlanConfigChange_DecodeArgs_Success(t *testing.T) {
	call := setupCallPlanConfigChange()
	assert.Equal(t, defaultPlanConfigChangeArgs, call.Args())

	buf := bytes.NewBuffer(somePlanConfigChangeArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, somePlanConfigChangeArgs, call.Args())
}

func Test_Call_PlanConfigChange_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{byte(moduleId), functionPlanConfigChangeIndex}, defaultPlanConfigChangeArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallPlanConfigChange()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_PlanConfigChange_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{byte(moduleId), functionPlanConfigChangeIndex}, somePlanConfigChangeArgs.Bytes()...))
	buf := bytes.NewBuffer(somePlanConfigChangeArgs.Bytes())

	call := setupCallPlanConfigChange()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_PlanConfigChange_Bytes(t *testing.T) {
	expected := append([]byte{byte(moduleId), functionPlanConfigChangeIndex}, defaultPlanConfigChangeArgs.Bytes()...)

	call := setupCallPlanConfigChange()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_PlanConfigChange_ModuleIndex(t *testing.T) {
	mockStoragePendingEpochConfigChange = new(mocks.StorageValue[NextConfigDescriptor])

	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallPlanConfigChange(tc, functionPlanConfigChangeIndex, dbWeight, mockStoragePendingEpochConfigChange)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_PlanConfigChange_FunctionIndex(t *testing.T) {
	mockStoragePendingEpochConfigChange = new(mocks.StorageValue[NextConfigDescriptor])

	testCases := []sc.U8{
		0,
		1,
		3,
		functionPlanConfigChangeIndex,
	}

	for _, tc := range testCases {
		call := newCallPlanConfigChange(moduleId, tc, dbWeight, mockStoragePendingEpochConfigChange)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_PlanConfigChange_BaseWeight(t *testing.T) {
	call := setupCallPlanConfigChange()

	assert.Equal(t, callPlanConfigChangeWeight(dbWeight), call.BaseWeight())
}

func Test_Call_PlanConfigChange_WeighData(t *testing.T) {
	call := setupCallPlanConfigChange()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_PlanConfigChange_ClassifyDispatch(t *testing.T) {
	call := setupCallPlanConfigChange()

	assert.Equal(t, primitives.NewDispatchClassNormal(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_PlanConfigChange_PaysFee(t *testing.T) {
	call := setupCallPlanConfigChange()

	assert.Equal(t, primitives.PaysYes, call.PaysFee(baseWeight))
}

func Test_Call_PlanConfigChange_Dispatch(t *testing.T) {
	call := setupCallPlanConfigChange()

	call, err := call.DecodeArgs(bytes.NewBuffer(somePlanConfigChangeArgs.Bytes()))
	assert.Nil(t, err)

	mockStoragePendingEpochConfigChange.On("Put", nextConfig).Return(nil)

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	assert.Nil(t, dispatchErr)

	mockStoragePendingEpochConfigChange.AssertCalled(t, "Put", nextConfig)
}

func setupCallPlanConfigChange() primitives.Call {
	mockStoragePendingEpochConfigChange = new(mocks.StorageValue[NextConfigDescriptor])

	return newCallPlanConfigChange(moduleId, functionPlanConfigChangeIndex, dbWeight, mockStoragePendingEpochConfigChange)
}
