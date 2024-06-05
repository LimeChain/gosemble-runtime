package system

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	keys = sc.Sequence[sc.Sequence[sc.U8]]{
		sc.BytesToSequenceU8([]byte("testkey1")),
		sc.BytesToSequenceU8([]byte("testkey2")),
	}

	someKillStorageArgs    = sc.NewVaryingData(keys)
	defaultKillStorageArgs = sc.NewVaryingData(sc.Sequence[sc.Sequence[sc.U8]]{})
)

func Test_Call_KillStorage_New(t *testing.T) {
	call := setupCallKillStorage()

	expected := callKillStorage{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionKillStorageIndex,
			Arguments:  defaultKillStorageArgs,
		},
		dbWeight:  dbWeight,
		ioStorage: mockIoStorage,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_KillStorage_DecodeArgs_Success(t *testing.T) {
	call := setupCallKillStorage()
	assert.Equal(t, defaultKillStorageArgs, call.Args())

	buf := bytes.NewBuffer(someKillStorageArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someKillStorageArgs, call.Args())
}

func Test_Call_KillStorage_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionKillStorageIndex}, defaultKillStorageArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallKillStorage()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_KillStorage_Encode_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionKillStorageIndex}, someKillStorageArgs.Bytes()...))
	buf := bytes.NewBuffer(someKillStorageArgs.Bytes())

	call := setupCallKillStorage()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_KillStorage_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionKillStorageIndex}, defaultKillStorageArgs.Bytes()...)

	call := setupCallKillStorage()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_KillStorage_ModuleIndex(t *testing.T) {
	initMockStorage()

	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallKillStorage(tc, functionKillStorageIndex, dbWeight, mockIoStorage)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_KillStorage_FunctionIndex(t *testing.T) {
	initMockStorage()

	testCases := []sc.U8{
		2,
		3,
		4,
		functionKillStorageIndex,
	}

	for _, tc := range testCases {
		call := newCallKillStorage(moduleId, tc, dbWeight, mockIoStorage)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_KillStorage_BaseWeight(t *testing.T) {
	call := setupCallKillStorage()

	assert.Equal(t, callKillStorageWeight(primitives.RuntimeDbWeight{}, 0), call.BaseWeight())
}

func Test_Call_KillStorage_WeighData(t *testing.T) {
	call := setupCallKillStorage()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_KillStorage_ClassifyDispatch(t *testing.T) {
	call := setupCallKillStorage()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_KillStorage_PaysFee(t *testing.T) {
	call := setupCallKillStorage()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_KillStorage_Dispatch(t *testing.T) {
	call := setupCallKillStorage()
	call, err := call.DecodeArgs(bytes.NewBuffer(someKillStorageArgs.Bytes()))
	assert.Nil(t, err)

	mockIoStorage.On("Clear", []byte("testkey1")).Return()
	mockIoStorage.On("Clear", []byte("testkey2")).Return()

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	assert.Nil(t, dispatchErr)
	mockIoStorage.AssertCalled(t, "Clear", []byte("testkey1"))
	mockIoStorage.AssertCalled(t, "Clear", []byte("testkey2"))
}

func setupCallKillStorage() primitives.Call {
	initMockStorage()
	return newCallKillStorage(moduleId, functionKillStorageIndex, dbWeight, mockIoStorage)
}
