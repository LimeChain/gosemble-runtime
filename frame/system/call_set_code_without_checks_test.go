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
	someSetCodeWithoutChecksArgs    = sc.NewVaryingData(codeBlob)
	defaultSetCodeWithoutChecksArgs = sc.NewVaryingData(sc.Sequence[sc.U8]{})
)

func Test_Call_SetCodeWithoutChecks_New(t *testing.T) {
	call := setupCallSetCodeWithoutChecks()

	expected := callSetCodeWithoutChecks{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionSetCodeWithoutChecksIndex,
			Arguments:  defaultSetCodeWithoutChecksArgs,
		},
		constants:     *moduleConstants,
		dbWeight:      dbWeight,
		hookOnSetCode: mockOnSetCode,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_SetCodeWithoutChecks_DecodeArgs_Success(t *testing.T) {
	call := setupCallSetCodeWithoutChecks()

	buf := bytes.NewBuffer(someSetCodeWithoutChecksArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someSetCodeWithoutChecksArgs, call.Args())
}

func Test_Call_SetCodeWithoutChecks_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionSetCodeWithoutChecksIndex}, defaultSetCodeWithoutChecksArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallSetCodeWithoutChecks()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_SetCodeWithoutChecks_Encode_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionSetCodeWithoutChecksIndex}, someSetCodeWithoutChecksArgs.Bytes()...))
	buf := bytes.NewBuffer(someSetCodeWithoutChecksArgs.Bytes())

	call := setupCallSetCodeWithoutChecks()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_SetCodeWithoutChecks_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSetCodeWithoutChecksIndex}, defaultSetCodeWithoutChecksArgs.Bytes()...)

	call := setupCallSetCodeWithoutChecks()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_SetCodeWithoutChecks_ModuleIndex(t *testing.T) {
	initMockStorage()
	mockOnSetCode := new(mocks.DefaultOnSetCode)

	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallSetCodeWithoutChecks(tc, functionSetCodeIndex, dbWeight, *moduleConstants, mockOnSetCode)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_SetCodeWithoutChecks_FunctionIndex(t *testing.T) {
	initMockStorage()
	mockOnSetCode := new(mocks.DefaultOnSetCode)

	testCases := []sc.U8{
		0,
		1,
		2,
		functionSetCodeWithoutChecksIndex,
	}

	for _, tc := range testCases {
		call := newCallSetCodeWithoutChecks(moduleId, tc, dbWeight, *moduleConstants, mockOnSetCode)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_SetCodeWithoutChecks_BaseWeight(t *testing.T) {
	call := setupCallSetCodeWithoutChecks()

	assert.Equal(t, callSetCodeWithoutChecksWeight(dbWeight), call.BaseWeight())
}

func Test_Call_SetCodeWithoutChecks_WeighData(t *testing.T) {
	call := setupCallSetCodeWithoutChecks()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_SetCodeWithoutChecks_ClassifyDispatch(t *testing.T) {
	call := setupCallSetCodeWithoutChecks()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_SetCodeWithoutChecks_PaysFee(t *testing.T) {
	call := setupCallSetCodeWithoutChecks()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_SetCodeWithoutChecks_Dispatch(t *testing.T) {
	call := setupCallSetCodeWithoutChecks()
	call, err := call.DecodeArgs(bytes.NewBuffer(someSetCodeWithoutChecksArgs.Bytes()))
	assert.Nil(t, err)

	mockOnSetCode.On("SetCode", codeBlob).Return(nil)

	res, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	mockOnSetCode.AssertCalled(t, "SetCode", codeBlob)

	assert.Nil(t, dispatchErr)
	assert.Equal(t, sc.NewOption[primitives.Weight](blockWeights.MaxBlock), res.ActualWeight)
}

func setupCallSetCodeWithoutChecks() primitives.Call {
	mockOnSetCode = new(mocks.DefaultOnSetCode)
	return newCallSetCodeWithoutChecks(moduleId, functionSetCodeWithoutChecksIndex, dbWeight, *moduleConstants, mockOnSetCode)
}
