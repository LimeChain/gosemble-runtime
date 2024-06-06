package system

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	addr, _    = primitives.NewAddress32(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	accountId  = primitives.NewAccountIdFromAddress32(addr)
	hashBytes  = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	msgHash, _ = primitives.NewH256(sc.BytesToFixedSequenceU8(hashBytes)...)
	event      = newEventRemarked(moduleId, accountId, msgHash)

	someRemarkWithEventArgs    = sc.NewVaryingData(remarkMsg)
	defaultRemarkWithEventArgs = sc.NewVaryingData(sc.Sequence[sc.U8]{})
)

var (
	mockEventDepositor *mocks.SystemModule
)

func Test_Call_RemarkWithEvent_New(t *testing.T) {
	call := setupCallRemarkWithEvent()

	expected := callRemarkWithEvent{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionRemarkWithEventIndex,
			Arguments:  defaultRemarkWithEventArgs,
		},
		dbWeight:       dbWeight,
		eventDepositor: mockEventDepositor,
		ioHashing:      mockIoHashing,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_RemarkWithEvent_DecodeArgs_Success(t *testing.T) {
	buf := bytes.NewBuffer(someRemarkWithEventArgs.Bytes())

	call := setupCallRemarkWithEvent()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	assert.Equal(t, someRemarkWithEventArgs, call.Args())
}

func Test_Call_RemarkWithEvent_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionRemarkWithEventIndex}, defaultRemarkWithEventArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallRemarkWithEvent()
	err := call.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuf, buf)
}

func Test_Call_RemarkWithEvent_EncodeWithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionRemarkWithEventIndex}, someRemarkWithEventArgs.Bytes()...))
	buf := bytes.NewBuffer(someRemarkWithEventArgs.Bytes())

	call := setupCallRemarkWithEvent()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuf, buf)
}

func Test_Call_RemarkWithEvent_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionRemarkWithEventIndex}, defaultRemarkWithEventArgs.Bytes()...)

	call := setupCallRemarkWithEvent()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_RemarkWithEvent_ModuleIndex(t *testing.T) {
	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallRemarkWithEvent(tc, functionRemarkWithEventIndex, dbWeight, mockIoHashing, mockEventDepositor)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_RemarkWithEvent_FunctionIndex(t *testing.T) {
	testCases := []sc.U8{
		1,
		2,
		3,
		functionRemarkWithEventIndex,
	}

	for _, tc := range testCases {
		call := newCallRemarkWithEvent(moduleId, tc, dbWeight, mockIoHashing, mockEventDepositor)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_RemarkWithEvent_BaseWeight(t *testing.T) {
	call := setupCallRemarkWithEvent()

	assert.Equal(t, callRemarkWithEventWeight(dbWeight, 0), call.BaseWeight())
}

func Test_Call_RemarkWithEvent_WeighData(t *testing.T) {
	call := setupCallRemarkWithEvent()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_RemarkWithEvent_ClassifyDispatch(t *testing.T) {
	call := setupCallRemarkWithEvent()

	assert.Equal(t, primitives.NewDispatchClassNormal(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_RemarkWithEvent_PaysFee(t *testing.T) {
	call := setupCallRemarkWithEvent()

	assert.Equal(t, primitives.PaysYes, call.PaysFee(baseWeight))
}

func Test_Call_RemarkWithEvent_Dispatch_Success(t *testing.T) {
	call := setupCallRemarkWithEvent()
	call, err := call.DecodeArgs(bytes.NewBuffer(someRemarkWithEventArgs.Bytes()))
	assert.Nil(t, err)

	mockIoHashing.On("Blake256", sc.SequenceU8ToBytes(remarkMsg)).Return(hashBytes)
	mockEventDepositor.On("DepositEvent", event).Return(nil)

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginSigned(accountId), call.Args())

	assert.Nil(t, dispatchErr)
	mockIoHashing.AssertCalled(t, "Blake256", sc.SequenceU8ToBytes(remarkMsg))
	mockEventDepositor.AssertCalled(t, "DepositEvent", event)
}

func setupCallRemarkWithEvent() primitives.Call {
	initMockStorage()
	mockEventDepositor = new(mocks.SystemModule)
	return newCallRemarkWithEvent(moduleId, functionRemarkWithEventIndex, dbWeight, mockIoHashing, mockEventDepositor)
}
