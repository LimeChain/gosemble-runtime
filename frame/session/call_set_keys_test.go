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
	setKeysArgs        = sc.NewVaryingData(sc.FixedSequence[primitives.Sr25519PublicKey]{sr25519PublicKey}, sc.Sequence[sc.U8]{})
	defaultSetKeysArgs = sc.NewVaryingData(sc.FixedSequence[primitives.Sr25519PublicKey]{}, sc.Sequence[sc.U8]{})
)

func Test_Call_SetKeys_New(t *testing.T) {
	target := setupCallSetKeys()

	expect := callSetKeys{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionSetKeys,
			Arguments:  defaultSetKeysArgs,
		},
		keyManager: mockKeyManager,
		handler:    mockSessionHandler,
		dbWeight:   dbWeight,
	}

	assert.Equal(t, expect, target)
}

func Test_Call_SetKeys_DecodeArgs(t *testing.T) {
	keys := sc.FixedSequence[primitives.Sr25519PublicKey]{sr25519PublicKey}
	buffer := bytes.NewBuffer([]byte{0})
	target := setupCallSetKeys()

	mockSessionHandler.On("DecodeKeys", buffer).Return(keys, nil)

	call, err := target.DecodeArgs(buffer)
	assert.Nil(t, err)
	assert.Equal(t, sc.NewVaryingData(keys, sc.Sequence[sc.U8]{}), call.Args())

	mockSessionHandler.AssertCalled(t, "DecodeKeys", buffer)
}

func Test_Call_SetKeys_Encode(t *testing.T) {
	target := setupCallSetKeys()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSetKeys}, defaultSetKeysArgs.Bytes()...))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_SetKeys_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSetKeys}, defaultSetKeysArgs.Bytes()...)

	target := setupCallSetKeys()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_SetKeys_ModuleIndex(t *testing.T) {
	target := setupCallSetKeys()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_SetKeys_FunctionIndex(t *testing.T) {
	target := setupCallSetKeys()

	assert.Equal(t, sc.U8(functionSetKeys), target.FunctionIndex())
}

func Test_Call_SetKeys_BaseWeight(t *testing.T) {
	target := setupCallSetKeys()

	assert.Equal(t, callSetKeysWeight(dbWeight), target.BaseWeight())
}

func Test_Call_SetKeys_WeighData(t *testing.T) {
	target := setupCallSetKeys()
	assert.Equal(t, primitives.WeightFromParts(234, 0), target.WeighData(baseWeight))
}

func Test_Call_SetKeys_ClassifyDispatch(t *testing.T) {
	target := setupCallSetKeys()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_SetKeys_PaysFee(t *testing.T) {
	target := setupCallSetKeys()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_SetKeys_Dispatch_Success(t *testing.T) {
	args := sc.NewVaryingData(sc.FixedSequence[primitives.Sr25519PublicKey]{sr25519PublicKey})
	target := setupCallSetKeys()

	mockSessionHandler.On("KeyTypeIds").Return(keyTypeIds)
	mockKeyManager.On("DoSetKeys", constants.OneAccountId, sessionKeys).Return(nil)

	res, err := target.Dispatch(primitives.NewRawOriginSigned(constants.OneAccountId), args)
	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)

	mockSessionHandler.AssertCalled(t, "KeyTypeIds")
	mockKeyManager.AssertCalled(t, "DoSetKeys", constants.OneAccountId, sessionKeys)
}

func Test_Call_SetKeys_Dispatch_InvalidOrigin(t *testing.T) {
	target := setupCallSetKeys()

	res, err := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData())

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)
}

func Test_Call_SetKeys_Docs(t *testing.T) {
	target := setupCallSetKeys()

	assert.Equal(t,
		"Sets the session key(s) of the function caller to `keys`. "+
			"Allows an account to set its session key prior to becoming a validator."+
			"This doesn't take effect until the next session.",
		target.Docs())
}

func setupCallSetKeys() primitives.Call {
	mockKeyManager = new(MockKeyManager)
	mockSessionHandler = new(MockSessionHandler)

	return newCallSetKeys(moduleId, functionSetKeys, dbWeight, mockKeyManager, mockSessionHandler)
}
