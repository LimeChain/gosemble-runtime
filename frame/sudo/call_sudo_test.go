package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Call_Sudo_ModuleIndex(t *testing.T) {
	target := setupCallSudo()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_Sudo_FunctionIndex(t *testing.T) {
	target := setupCallSudo()

	assert.Equal(t, sc.U8(functionSudo), target.FunctionIndex())
}

func Test_Call_Sudo_BaseWeight(t *testing.T) {
	target := setupCallSudo()

	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(baseWeight)

	assert.Equal(t, callSudoWeight(dbWeight).Add(baseWeight), target.BaseWeight())

	mockCall.AssertCalled(t, "BaseWeight")
	mockCall.AssertCalled(t, "WeighData", baseWeight)
}

func Test_Call_Sudo_WeighData(t *testing.T) {
	target := setupCallSudo()
	assert.Equal(t, primitives.WeightFromParts(234, 0), target.WeighData(baseWeight))
}

func Test_Call_Sudo_ClassifyDispatch(t *testing.T) {
	target := setupCallSudo()

	mockCall.On("ClassifyDispatch", baseWeight).Return(primitives.NewDispatchClassNormal())

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))

	mockCall.AssertCalled(t, "ClassifyDispatch", baseWeight)
}

func Test_Call_Sudo_PaysFee(t *testing.T) {
	target := setupCallSudo()

	assert.Equal(t, primitives.PaysNo, target.PaysFee(baseWeight))
}

func Test_Call_Sudo_Encode(t *testing.T) {
	target := setupCallSudo()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSudo}))
	buf := &bytes.Buffer{}

	mockCall.On("Encode", expectedBuffer).Return(nil)

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)

	mockCall.AssertCalled(t, "Encode", expectedBuffer)
}

func Test_Call_Sudo_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSudo})
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSudo}))
	target := setupCallSudo()

	mockCall.On("Encode", expectedBuffer).Return(nil)

	assert.Equal(t, expected, target.Bytes())

	mockCall.AssertCalled(t, "Encode", expectedBuffer)
}

func Test_Call_Sudo_DecodeSudoArgs(t *testing.T) {
	buffer := &bytes.Buffer{}

	target := setupCallSudo()

	call, err := target.DecodeSudoArgs(buffer, func(buffer *bytes.Buffer) (primitives.Call, error) { return mockCall, nil })
	assert.Nil(t, err)
	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, sc.NewVaryingData(mockCall), call.Args())
}

func Test_Call_Sudo_DecodeArgs(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{1})
	target := setupCallSudo()

	assert.PanicsWithValue(t, "not implemented", func() {
		target.DecodeArgs(buffer)
	})
	assert.Equal(t, 1, buffer.Len())
}

func Test_Call_Sudo_Docs(t *testing.T) {
	target := setupCallSudo()

	assert.Equal(t, "Authenticates the sudo key and dispatches a call function with `Root` origin.", target.Docs())
}

func Test_Call_Sudo_Dispatch(t *testing.T) {
	target := setupCallSudo()

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

func Test_Call_Sudo_Dispatch_Fails_InvalidOrigin(t *testing.T) {
	target := setupCallSudo()

	res, err := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData())
	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)
}

func setupCallSudo() callSudo {
	module := setupModule()

	call := newCallSudo(moduleId, functionSudo, dbWeight, module).(callSudo)
	call.Arguments = sc.NewVaryingData(mockCall)

	return call
}
