package system

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	codeBlob           = sc.Sequence[sc.U8]{1, 2, 3}
	someSetCodeArgs    = sc.NewVaryingData(codeBlob)
	defaultSetCodeArgs = sc.NewVaryingData(sc.Sequence[sc.U8]{})
)

var (
	mockCodeUpgrader *mocks.SystemModule
	mockOnSetCode    *mocks.DefaultOnSetCode
)

func Test_Call_SetCode_New(t *testing.T) {
	call := setupCallSetCode()

	expected := callSetCode{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionSetCodeIndex,
			Arguments:  defaultSetCodeArgs,
		},
		constants:     *moduleConstants,
		dbWeight:      dbWeight,
		hookOnSetCode: mockOnSetCode,
		codeUpgrader:  mockCodeUpgrader,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_SetCode_DecodeArgs_Success(t *testing.T) {
	call := setupCallSetCode()

	buf := bytes.NewBuffer(someSetCodeArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someSetCodeArgs, call.Args())
}

func Test_Call_SetCode_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionSetCodeIndex}, defaultSetCodeArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallSetCode()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_SetCode_Encode_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionSetCodeIndex}, someSetCodeArgs.Bytes()...))

	buf := bytes.NewBuffer(someSetCodeArgs.Bytes())

	call := setupCallSetCode()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_SetCode_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSetCodeIndex}, defaultSetCodeArgs.Bytes()...)

	call := setupCallSetCode()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_SetCode_ModuleIndex(t *testing.T) {
	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallSetCode(tc, functionSetCodeIndex, dbWeight, *moduleConstants, mockOnSetCode, mockCodeUpgrader)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_SetCode_FunctionIndex(t *testing.T) {
	testCases := []sc.U8{
		0,
		1,
		functionSetCodeIndex,
		3,
	}

	for _, tc := range testCases {
		call := newCallSetCode(moduleId, tc, dbWeight, *moduleConstants, mockOnSetCode, mockCodeUpgrader)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_SetCode_BaseWeight(t *testing.T) {
	call := setupCallSetCode()

	assert.Equal(t, callSetCodeWeight(dbWeight), call.BaseWeight())
}

func Test_Call_SetCode_WeighData(t *testing.T) {
	call := setupCallSetCode()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_SetCode_ClassifyDispatch(t *testing.T) {
	call := setupCallSetCode()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_SetCode_PaysFee(t *testing.T) {
	call := setupCallSetCode()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_SetCode_Dispatch_Error_Failed_To_Extract_Runtime_Version(t *testing.T) {
	call := setupCallSetCode()
	call, err := call.DecodeArgs(bytes.NewBuffer(someSetCodeArgs.Bytes()))
	assert.Nil(t, err)

	mockCodeUpgrader.On("CanSetCode", codeBlob).Return(NewDispatchErrorFailedToExtractRuntimeVersion(moduleId))

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	mockCodeUpgrader.AssertCalled(t, "CanSetCode", codeBlob)
	mockOnSetCode.AssertNotCalled(t, "SetCode", codeBlob)
	assert.Equal(t, NewDispatchErrorFailedToExtractRuntimeVersion(moduleId), dispatchErr)
}

func Test_Call_SetCode_Dispatch_Error_Invalid_Spec_Name(t *testing.T) {
	call := setupCallSetCode()
	call, err := call.DecodeArgs(bytes.NewBuffer(someSetCodeArgs.Bytes()))
	assert.Nil(t, err)

	mockCodeUpgrader.On("CanSetCode", codeBlob).Return(NewDispatchErrorInvalidSpecName(moduleId))

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	mockCodeUpgrader.AssertCalled(t, "CanSetCode", codeBlob)
	mockOnSetCode.AssertNotCalled(t, "SetCode", codeBlob)
	assert.Equal(t, NewDispatchErrorInvalidSpecName(moduleId), dispatchErr)
}

func Test_Call_SetCode_Dispatch_Error_Spec_Version_Needs_To_Increase(t *testing.T) {
	call := setupCallSetCode()
	call, err := call.DecodeArgs(bytes.NewBuffer(someSetCodeArgs.Bytes()))
	assert.Nil(t, err)

	mockCodeUpgrader.On("CanSetCode", codeBlob).Return(NewDispatchErrorSpecVersionNeedsToIncrease(moduleId))

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	mockCodeUpgrader.AssertCalled(t, "CanSetCode", codeBlob)
	mockOnSetCode.AssertNotCalled(t, "SetCode", codeBlob)
	assert.Equal(t, NewDispatchErrorSpecVersionNeedsToIncrease(moduleId), dispatchErr)
}

func Test_Call_SetCode_Dispatch(t *testing.T) {
	call := setupCallSetCode()
	call, err := call.DecodeArgs(bytes.NewBuffer(someSetCodeArgs.Bytes()))
	assert.Nil(t, err)

	mockCodeUpgrader.On("CanSetCode", codeBlob).Return(nil)
	mockOnSetCode.On("SetCode", codeBlob).Return(nil)

	res, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())
	mockCodeUpgrader.AssertCalled(t, "CanSetCode", codeBlob)
	mockOnSetCode.AssertCalled(t, "SetCode", codeBlob)

	assert.Nil(t, dispatchErr)
	assert.Equal(t, sc.NewOption[primitives.Weight](blockWeights.MaxBlock), res.ActualWeight)
}

func setupCallSetCode() primitives.Call {
	mockCodeUpgrader = new(mocks.SystemModule)
	mockOnSetCode = new(mocks.DefaultOnSetCode)
	return newCallSetCode(moduleId, functionSetCodeIndex, dbWeight, *moduleConstants, mockOnSetCode, mockCodeUpgrader)
}
