package session

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

var (
	mockKeyManager *MockKeyManager
)

func Test_Call_PurgeKeys_New(t *testing.T) {
	target := setupCallPurgeKeys()

	expect := callPurgeKeys{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionPurgeKeys,
			Arguments:  sc.NewVaryingData(),
		},
		keyManager: mockKeyManager,
		dbWeight:   dbWeight,
	}

	assert.Equal(t, expect, target)
}

func Test_Call_PurgeKeys_DecodeArgs(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{1})
	target := setupCallPurgeKeys()

	call, err := target.DecodeArgs(buffer)
	assert.Nil(t, err)

	assert.Equal(t, 1, buffer.Len())
	assert.Equal(t, sc.NewVaryingData(), call.Args())
}

func Test_Call_PurgeKeys_Encode(t *testing.T) {
	target := setupCallPurgeKeys()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionPurgeKeys}))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_PurgeKeys_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionPurgeKeys})

	target := setupCallPurgeKeys()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_PurgeKeys_ModuleIndex(t *testing.T) {
	target := setupCallPurgeKeys()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_PurgeKeys_FunctionIndex(t *testing.T) {
	target := setupCallPurgeKeys()

	assert.Equal(t, sc.U8(functionPurgeKeys), target.FunctionIndex())
}

func Test_Call_PurgeKeys_BaseWeight(t *testing.T) {
	target := setupCallPurgeKeys()

	assert.Equal(t, callPurgeKeysWeight(dbWeight), target.BaseWeight())
}

func Test_Call_PurgeKeys_WeighData(t *testing.T) {
	target := setupCallPurgeKeys()
	assert.Equal(t, primitives.WeightFromParts(234, 0), target.WeighData(baseWeight))
}

func Test_Call_PurgeKeys_ClassifyDispatch(t *testing.T) {
	target := setupCallPurgeKeys()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_PurgeKeys_PaysFee(t *testing.T) {
	target := setupCallPurgeKeys()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_PurgeKeys_Dispatch_Success(t *testing.T) {
	target := setupCallPurgeKeys()

	mockKeyManager.On("DoPurgeKeys", constants.OneAccountId).Return(nil)

	res, err := target.Dispatch(primitives.NewRawOriginSigned(constants.OneAccountId), sc.NewVaryingData())
	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)

	mockKeyManager.AssertCalled(t, "DoPurgeKeys", constants.OneAccountId)
}

func Test_Call_PurgeKeys_Dispatch_InvalidOrigin(t *testing.T) {
	target := setupCallPurgeKeys()

	res, err := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData())

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)
}

func Test_Call_PurgeKeys_Docs(t *testing.T) {
	target := setupCallPurgeKeys()

	assert.Equal(t, "Removes any session key(s) of the function caller.", target.Docs())
}

func setupCallPurgeKeys() primitives.Call {
	mockKeyManager = new(MockKeyManager)

	return newCallPurgeKeys(moduleId, functionPurgeKeys, dbWeight, mockKeyManager)
}
