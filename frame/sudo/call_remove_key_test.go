package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	baseWeight = primitives.WeightFromParts(234, 456)
)

func Test_Call_RemoveKey_ModuleIndex(t *testing.T) {
	target := setupCallRemoveKey()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_RemoveKey_FunctionIndex(t *testing.T) {
	target := setupCallRemoveKey()

	assert.Equal(t, sc.U8(functionRemoveKey), target.FunctionIndex())
}

func Test_Call_RemoveKey_BaseWeight(t *testing.T) {
	target := setupCallRemoveKey()

	assert.Equal(t, callRemoveKeyWeight(dbWeight), target.BaseWeight())
}

func Test_Call_RemoveKey_WeighData(t *testing.T) {
	target := setupCallRemoveKey()
	assert.Equal(t, primitives.WeightFromParts(234, 0), target.WeighData(baseWeight))
}

func Test_Call_RemoveKey_ClassifyDispatch(t *testing.T) {
	target := setupCallRemoveKey()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_RemoveKey_PaysFee(t *testing.T) {
	target := setupCallRemoveKey()

	assert.Equal(t, primitives.PaysNo, target.PaysFee(baseWeight))
}

func Test_Call_RemoveKey_Encode(t *testing.T) {
	target := setupCallRemoveKey()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionRemoveKey}))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_RemoveKey_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionRemoveKey})

	target := setupCallRemoveKey()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_RemoveKey_DecodeSudoArgs(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{1})
	target := setupCallRemoveKey()

	call, err := target.DecodeSudoArgs(buffer, func(buffer *bytes.Buffer) (primitives.Call, error) { return mockCall, nil })
	assert.Nil(t, err)
	assert.Equal(t, 1, buffer.Len())
	assert.Equal(t, sc.NewVaryingData(), call.Args())
}

func Test_Call_RemoveKey_DecodeArgs(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{1})
	target := setupCallRemoveKey()

	assert.PanicsWithValue(t, "not implemented", func() {
		target.DecodeArgs(buffer)
	})
}

func Test_Call_RemoveKey_Docs(t *testing.T) {
	target := setupCallRemoveKey()

	assert.Equal(t, "Permanently removes the sudo key. This cannot be undone.", target.Docs())
}

func Test_Call_RemoveKey_Dispatch(t *testing.T) {
	target := setupCallRemoveKey()

	mockStorageKey.On("Get").Return(constants.OneAccountId, nil)
	mockEventDepositor.On("DepositEvent", newEventKeyRemoved(moduleId)).Return()
	mockStorageKey.On("Clear").Return()

	res, err := target.Dispatch(signedOrigin, sc.NewVaryingData())
	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{
		ActualWeight: sc.Option[primitives.Weight]{},
		PaysFee:      primitives.PaysNo,
	}, res)

	mockStorageKey.AssertCalled(t, "Get")
	mockEventDepositor.AssertCalled(t, "DepositEvent", newEventKeyRemoved(moduleId))
	mockStorageKey.AssertCalled(t, "Clear")
}

func Test_Call_RemoveKey_Dispatch_Fails_InvalidOrigin(t *testing.T) {
	target := setupCallRemoveKey()

	res, err := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData())
	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)
}

func setupCallRemoveKey() callRemoveKey {
	module := setupModule()

	return newCallRemoveKey(moduleId, functionRemoveKey, dbWeight, module).(callRemoveKey)
}
