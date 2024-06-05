package system

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	someSetStorageArgs = sc.NewVaryingData(
		sc.Sequence[KeyValue]{
			{
				Key:   sc.BytesToSequenceU8([]byte("testkey1")),
				Value: sc.BytesToSequenceU8([]byte("testvalue1")),
			},
			{
				Key:   sc.BytesToSequenceU8([]byte("testkey2")),
				Value: sc.BytesToSequenceU8([]byte("testvalue2")),
			},
		},
	)
	defaultSetStorageArgs = sc.NewVaryingData(sc.Sequence[KeyValue]{})
)

func Test_Call_SetStorage_New(t *testing.T) {
	call := setupCallSetStorage()

	expected := callSetStorage{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionSetStorageIndex,
			Arguments:  defaultSetStorageArgs,
		},
		dbWeight:  dbWeight,
		ioStorage: mockIoStorage,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_SetStorage_DecodeArgs_Success(t *testing.T) {
	call := setupCallSetStorage()
	assert.Equal(t, defaultSetStorageArgs, call.Args())

	buf := bytes.NewBuffer(someSetStorageArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someSetStorageArgs, call.Args())
}

func Test_Call_SetStorage_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionSetStorageIndex}, defaultSetStorageArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallSetStorage()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_SetStorage_Encode_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionSetStorageIndex}, someSetStorageArgs.Bytes()...))
	buf := bytes.NewBuffer(someSetStorageArgs.Bytes())

	call := setupCallSetStorage()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_SetStorage_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSetStorageIndex}, defaultSetStorageArgs.Bytes()...)

	call := setupCallSetStorage()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_SetStorage_ModuleIndex(t *testing.T) {
	initMockStorage()

	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallSetStorage(tc, functionSetStorageIndex, dbWeight, mockIoStorage)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_SetStorage_FunctionIndex(t *testing.T) {
	initMockStorage()

	testCases := []sc.U8{
		0,
		1,
		3,
		functionSetStorageIndex,
	}

	for _, tc := range testCases {
		call := newCallSetStorage(moduleId, tc, dbWeight, mockIoStorage)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_SetStorage_BaseWeight(t *testing.T) {
	call := setupCallSetStorage()

	assert.Equal(t, callSetStorageWeight(primitives.RuntimeDbWeight{}, 0), call.BaseWeight())
}

func Test_Call_SetStorage_WeighData(t *testing.T) {
	call := setupCallSetStorage()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_SetStorage_ClassifyDispatch(t *testing.T) {
	call := setupCallSetStorage()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_SetStorage_PaysFee(t *testing.T) {
	call := setupCallSetStorage()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_SetStorage_Dispatch(t *testing.T) {
	call := setupCallSetStorage()

	call, err := call.DecodeArgs(bytes.NewBuffer(someSetStorageArgs.Bytes()))
	assert.Nil(t, err)

	mockIoStorage.On("Set", []byte("testkey1"), []byte("testvalue1")).Return()
	mockIoStorage.On("Set", []byte("testkey2"), []byte("testvalue2")).Return()

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	assert.Nil(t, dispatchErr)
	mockIoStorage.AssertCalled(t, "Set", []byte("testkey1"), []byte("testvalue1"))
	mockIoStorage.AssertCalled(t, "Set", []byte("testkey2"), []byte("testvalue2"))
}

func setupCallSetStorage() primitives.Call {
	initMockStorage()
	return newCallSetStorage(moduleId, functionSetStorageIndex, dbWeight, mockIoStorage)
}
