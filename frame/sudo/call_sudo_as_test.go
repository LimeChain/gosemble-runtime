package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Call_SudoAs_ModuleIndex(t *testing.T) {
	target := setupCallSudoAs()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_SudoAs_FunctionIndex(t *testing.T) {
	target := setupCallSudoAs()

	assert.Equal(t, sc.U8(functionSudoAs), target.FunctionIndex())
}

func Test_Call_SudoAs_BaseWeight(t *testing.T) {
	target := setupCallSudoAs()

	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(baseWeight)

	assert.Equal(t, callSudoAsWeight(dbWeight).Add(baseWeight), target.BaseWeight())

	mockCall.AssertCalled(t, "BaseWeight")
	mockCall.AssertCalled(t, "WeighData", baseWeight)
}

func Test_Call_SudoAs_WeighData(t *testing.T) {
	target := setupCallSudoAs()
	assert.Equal(t, primitives.WeightFromParts(234, 0), target.WeighData(baseWeight))
}

func Test_Call_SudoAs_ClassifyDispatch(t *testing.T) {
	target := setupCallSudoAs()

	mockCall.On("ClassifyDispatch", baseWeight).Return(primitives.NewDispatchClassNormal())

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))

	mockCall.AssertCalled(t, "ClassifyDispatch", baseWeight)
}

func Test_Call_SudoAs_PaysFee(t *testing.T) {
	target := setupCallSudoAs()

	assert.Equal(t, primitives.PaysNo, target.PaysFee(baseWeight))
}

func Test_Call_SudoAs_Encode(t *testing.T) {
	target := setupCallSudoAs()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSudoAs}, newMultiAddress.Bytes()...))
	buf := &bytes.Buffer{}

	mockCall.On("Encode", expectedBuffer).Return(nil)

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)

	mockCall.AssertCalled(t, "Encode", expectedBuffer)
}

func Test_Call_SudoAs_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSudoAs}, newMultiAddress.Bytes()...)
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSudoAs}, newMultiAddress.Bytes()...))
	target := setupCallSudoAs()

	mockCall.On("Encode", expectedBuffer).Return(nil)

	assert.Equal(t, expected, target.Bytes())

	mockCall.AssertCalled(t, "Encode", expectedBuffer)
}

func Test_Call_SudoAs_DecodeSudoArgs(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(newMultiAddress.Bytes())

	target := setupCallSudoAs()

	call, err := target.DecodeSudoArgs(buffer, func(buffer *bytes.Buffer) (primitives.Call, error) { return mockCall, nil })
	assert.Nil(t, err)
	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, sc.NewVaryingData(newMultiAddress, mockCall), call.Args())
}

func Test_Call_SudoAs_DecodeArgs(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{1})
	target := setupCallSudoAs()

	assert.PanicsWithValue(t, "not implemented", func() {
		target.DecodeArgs(buffer)
	})
	assert.Equal(t, 1, buffer.Len())
}

func Test_Call_SudoAs_Dispatch(t *testing.T) {
	target := setupCallSudoAs()

	mockStorageKey.On("Get").Return(newKey, nil)
	mockCall.On("Args").Return(sc.NewVaryingData())
	mockCall.On("Dispatch", signedOrigin, sc.NewVaryingData()).Return(primitives.PostDispatchInfo{}, nil)
	mockEventDepositor.On("DepositEvent", newEventSudoAsDone(moduleId, dispatchOutcomeEmpty)).Return()

	res, err := target.Dispatch(signedOrigin, sc.NewVaryingData(newMultiAddress, mockCall))
	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{
		PaysFee: primitives.PaysNo,
	}, res)

	mockCall.AssertCalled(t, "Args")
	mockCall.AssertCalled(t, "Dispatch", signedOrigin, sc.NewVaryingData())
	mockEventDepositor.AssertCalled(t, "DepositEvent", newEventSudoAsDone(moduleId, dispatchOutcomeEmpty))
}

func Test_Call_SudoAs_Dispatch_Fails_InvalidOrigin(t *testing.T) {
	target := setupCallSudoAs()

	res, err := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData())
	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)
}

func Test_Call_SudoAs_Docs(t *testing.T) {
	target := setupCallSudoAs()

	assert.Equal(t, "Authenticates the sudo key and dispatches a function call with `Signed` origin from a given account."+
		"The dispatch origin for this call must be `Signed`.", target.Docs())
}

func setupCallSudoAs() callSudoAs {
	module := setupModule()

	call := newCallSudoAs(moduleId, functionSudoAs, dbWeight, module).(callSudoAs)
	call.Arguments = sc.NewVaryingData(newMultiAddress, mockCall)

	return call
}
