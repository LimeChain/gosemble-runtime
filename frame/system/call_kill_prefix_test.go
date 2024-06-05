package system

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	prefixBytes = []byte("prefixkey")
	prefix      = sc.Sequence[sc.U8](sc.BytesToSequenceU8(prefixBytes))
	subkeys     = sc.U32(1)
)

var (
	someKillPrefixArgs    = sc.NewVaryingData(prefix, subkeys)
	defaultKillPrefixArgs = sc.NewVaryingData(sc.Sequence[sc.U8]{}, sc.U32(0))
)

func Test_Call_KillPrefix_New(t *testing.T) {
	call := setupCallKillPrefix()

	expected := callKillPrefix{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionKillPrefixIndex,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}, sc.U32(0)),
		},
		dbWeight:  dbWeight,
		ioStorage: mockIoStorage,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_KillPrefix_DecodeArgs_Success(t *testing.T) {
	call := setupCallKillPrefix()
	assert.Equal(t, defaultKillPrefixArgs, call.Args())

	buf := bytes.NewBuffer(someKillPrefixArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someKillPrefixArgs, call.Args())
}

func Test_Call_KillPrefix_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionKillPrefixIndex}, defaultKillPrefixArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallKillPrefix()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_KillPrefix_Encode_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionKillPrefixIndex}, someKillPrefixArgs.Bytes()...))
	buf := bytes.NewBuffer(someKillPrefixArgs.Bytes())

	call := setupCallKillPrefix()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_KillPrefix_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionKillPrefixIndex}, defaultKillPrefixArgs.Bytes()...)

	call := setupCallKillPrefix()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_KillPrefix_ModuleIndex(t *testing.T) {
	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallKillPrefix(tc, functionKillPrefixIndex, dbWeight, mockIoStorage)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_KillPrefix_FunctionIndex(t *testing.T) {
	testCases := []sc.U8{
		3,
		4,
		5,
		functionKillPrefixIndex,
	}

	for _, tc := range testCases {
		call := newCallKillPrefix(moduleId, tc, dbWeight, mockIoStorage)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_KillPrefix_BaseWeight(t *testing.T) {
	call := setupCallKillPrefix()

	assert.Equal(t, callKillPrefixWeight(primitives.RuntimeDbWeight{}, 0), call.BaseWeight())
}

func Test_Call_KillPrefix_WeighData(t *testing.T) {
	call := setupCallKillPrefix()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_KillPrefix_ClassifyDispatch(t *testing.T) {
	call := setupCallKillPrefix()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_KillPrefix_PaysFee(t *testing.T) {
	call := setupCallKillPrefix()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_KillPrefix_Dispatch(t *testing.T) {
	call := setupCallKillPrefix()
	call, err := call.DecodeArgs(bytes.NewBuffer(someKillPrefixArgs.Bytes()))
	assert.Nil(t, err)

	mockIoStorage.On("ClearPrefix", prefixBytes, sc.NewOption[sc.U32](subkeys).Bytes()).Return()

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	assert.Nil(t, dispatchErr)
	mockIoStorage.AssertCalled(t, "ClearPrefix", prefixBytes, sc.NewOption[sc.U32](subkeys).Bytes())
}

func setupCallKillPrefix() primitives.Call {
	initMockStorage()
	return newCallKillPrefix(moduleId, functionKillPrefixIndex, dbWeight, mockIoStorage)
}
