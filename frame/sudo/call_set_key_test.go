package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func Test_Call_SetKey_ModuleIndex(t *testing.T) {
	target := setupCallSetKey()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_SetKey_FunctionIndex(t *testing.T) {
	target := setupCallSetKey()

	assert.Equal(t, sc.U8(functionSetKey), target.FunctionIndex())
}

func Test_Call_SetKey_BaseWeight(t *testing.T) {
	target := setupCallSetKey()

	assert.Equal(t, callSetKeyWeight(dbWeight), target.BaseWeight())
}

func Test_Call_SetKey_WeighData(t *testing.T) {
	target := setupCallSetKey()
	assert.Equal(t, primitives.WeightFromParts(234, 0), target.WeighData(baseWeight))
}

func Test_Call_SetKey_ClassifyDispatch(t *testing.T) {
	target := setupCallSetKey()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_SetKey_PaysFee(t *testing.T) {
	target := setupCallSetKey()

	assert.Equal(t, primitives.PaysNo, target.PaysFee(baseWeight))
}

func Test_Call_SetKey_Encode(t *testing.T) {
	target := setupCallSetKey()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSetKey}))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_SetKey_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSetKey})

	target := setupCallSetKey()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_SetKey_DecodeSudoArgs(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(newMultiAddress.Bytes())

	target := setupCallSetKey()

	call, err := target.DecodeSudoArgs(buffer, func(buffer *bytes.Buffer) (primitives.Call, error) { return mockCall, nil })
	assert.Nil(t, err)
	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, sc.NewVaryingData(newMultiAddress), call.Args())
}

func Test_Call_SetKey_DecodeSudoArgs_Fails_EmptyBuffer(t *testing.T) {
	buffer := &bytes.Buffer{}

	target := setupCallSetKey()

	call, err := target.DecodeSudoArgs(buffer, func(buffer *bytes.Buffer) (primitives.Call, error) { return mockCall, nil })
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, call)
}

func Test_Call_SetKey_DecodeArgs(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{1})
	target := setupCallSetKey()

	assert.PanicsWithValue(t, "not implemented", func() {
		target.DecodeArgs(buffer)
	})
	assert.Equal(t, 1, buffer.Len())
}

func Test_Call_SetKey_Docs(t *testing.T) {
	target := setupCallSetKey()

	assert.Equal(t, "Authenticates the current sudo key and sets the given `AccountId` as the new sudo.", target.Docs())
}

func Test_Call_SetKey_Dispatch_Fails_InvalidOrigin(t *testing.T) {
	target := setupCallSetKey()

	res, err := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData())
	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)
}

func Test_Call_SetKey_Dispatch(t *testing.T) {
	target := setupCallSetKey()

	mockStorageKey.On("Get").Return(oldKey, nil)
	mockEventDepositor.On("DepositEvent", newEventKeyChanged(moduleId, sc.NewOption[primitives.AccountId](oldKey), newKey)).Return()
	mockStorageKey.On("Put", newKey).Return()

	res, err := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(newMultiAddress))
	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{
		PaysFee: primitives.PaysNo,
	}, res)

	mockStorageKey.AssertCalled(t, "Get")
	mockEventDepositor.AssertCalled(t, "DepositEvent", newEventKeyChanged(moduleId, sc.NewOption[primitives.AccountId](oldKey), newKey))
	mockStorageKey.AssertCalled(t, "Put", newKey)
}

func setupCallSetKey() callSetKey {
	module := setupModule()

	return newCallSetKey(moduleId, functionSetKey, dbWeight, module).(callSetKey)
}
