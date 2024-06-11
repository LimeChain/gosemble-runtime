package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	weight = primitives.WeightFromParts(7, 8)
)

func Test_Call_SudoUncheckedWeight_ModuleIndex(t *testing.T) {
	target := setupCallSudoUncheckedWeight()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_SudoUncheckedWeight_FunctionIndex(t *testing.T) {
	target := setupCallSudoUncheckedWeight()

	assert.Equal(t, sc.U8(functionSudoUncheckedWeight), target.FunctionIndex())
}

func Test_Call_SudoUncheckedWeight_BaseWeight(t *testing.T) {
	target := setupCallSudoUncheckedWeight()

	assert.Equal(t, weight, target.BaseWeight())
}

func Test_Call_SudoUncheckedWeight_InvalidArgument(t *testing.T) {
	target := setupCallSudoUncheckedWeight()
	target.Arguments = sc.NewVaryingData(weight, mockCall)

	assert.PanicsWithValue(t, "invalid [1] argument Weight", func() {
		target.BaseWeight()
	})
}

func Test_Call_SudoUncheckedWeight_WeighData(t *testing.T) {
	target := setupCallSudoUncheckedWeight()
	assert.Equal(t, primitives.WeightFromParts(234, 0), target.WeighData(baseWeight))
}

func Test_Call_SudoUncheckedWeight_ClassifyDispatch(t *testing.T) {
	target := setupCallSudoUncheckedWeight()

	mockCall.On("ClassifyDispatch", baseWeight).Return(primitives.NewDispatchClassNormal())

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))

	mockCall.AssertCalled(t, "ClassifyDispatch", baseWeight)
}

func Test_Call_SudoUncheckedWeight_PaysFee(t *testing.T) {
	target := setupCallSudoUncheckedWeight()

	assert.Equal(t, primitives.PaysNo, target.PaysFee(baseWeight))
}

func Test_Call_SudoUncheckedWeight_Encode(t *testing.T) {
	target := setupCallSudoUncheckedWeight()
	callBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSudoUncheckedWeight}))
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSudoUncheckedWeight}, weight.Bytes()...))
	buf := &bytes.Buffer{}

	mockCall.On("Encode", callBuffer).Return(nil)

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)

	mockCall.AssertCalled(t, "Encode", expectedBuffer)
}

func Test_Call_SudoUncheckedWeight_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSudoUncheckedWeight}, weight.Bytes()...)
	callBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSudoUncheckedWeight}))
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSudoUncheckedWeight}, weight.Bytes()...))
	target := setupCallSudoUncheckedWeight()

	mockCall.On("Encode", callBuffer).Return(nil)

	assert.Equal(t, expected, target.Bytes())

	mockCall.AssertCalled(t, "Encode", expectedBuffer)
}

func Test_Call_SudoUncheckedWeight_DecodeSudoArgs(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(weight.Bytes())

	target := setupCallSudoUncheckedWeight()

	call, err := target.DecodeSudoArgs(buffer, func(buffer *bytes.Buffer) (primitives.Call, error) { return mockCall, nil })
	assert.Nil(t, err)
	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, sc.NewVaryingData(mockCall, weight), call.Args())
}

func Test_Call_SudoUncheckedWeight_DecodeArgs(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{1})
	target := setupCallSudoUncheckedWeight()

	assert.PanicsWithValue(t, "not implemented", func() {
		target.DecodeArgs(buffer)
	})
	assert.Equal(t, 1, buffer.Len())
}

func Test_Call_SudoUncheckedWeight_Dispatch_Fails_InvalidOrigin(t *testing.T) {
	target := setupCallSudoUncheckedWeight()

	res, err := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData())
	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)
}

func Test_Call_SudoUncheckedWeight_Dispatch(t *testing.T) {
	target := setupCallSudoUncheckedWeight()

	mockStorageKey.On("Get").Return(newKey, nil)
	mockCall.On("Args").Return(sc.NewVaryingData())
	mockCall.On("Dispatch", primitives.NewRawOriginRoot(), sc.NewVaryingData()).Return(primitives.PostDispatchInfo{}, nil)
	mockEventDepositor.On("DepositEvent", newEventSudid(moduleId, dispatchOutcomeEmpty)).Return()

	res, err := target.Dispatch(signedOrigin, sc.NewVaryingData(mockCall))
	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{
		PaysFee: primitives.PaysNo,
	}, res)

	mockCall.AssertCalled(t, "Args")
	mockCall.AssertCalled(t, "Dispatch", primitives.NewRawOriginRoot(), sc.NewVaryingData())
	mockEventDepositor.AssertCalled(t, "DepositEvent", newEventSudid(moduleId, dispatchOutcomeEmpty))
}

func Test_Call_SudoUncheckedWeight_Docs(t *testing.T) {
	target := setupCallSudoUncheckedWeight()

	assert.Equal(t, docsCallSudoUncheckedWeight, target.Docs())
}

func setupCallSudoUncheckedWeight() callSudoUncheckedWeight {
	module := setupModule()

	call := newCallSudoUncheckedWeight(moduleId, functionSudoUncheckedWeight, dbWeight, module).(callSudoUncheckedWeight)
	call.Arguments = sc.NewVaryingData(mockCall, weight)

	return call
}
