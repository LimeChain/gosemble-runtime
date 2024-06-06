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
	digestItem              = primitives.NewDigestItemRuntimeEnvironmentUpgrade()
	pages                   = sc.U64(123)
	someSetHeapPagesArgs    = sc.NewVaryingData(pages)
	defaultSetHeapPagesArgs = sc.NewVaryingData(sc.U64(0))
)

var (
	mockLogDepositor *mocks.SystemModule
)

func Test_Call_SetHeapPages_New(t *testing.T) {
	call := setupCallSetHeapPages()

	expected := callSetHeapPages{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionSetHeapPagesIndex,
			Arguments:  defaultSetHeapPagesArgs,
		},
		dbWeight:     dbWeight,
		heapPages:    mockStorageHeapPages,
		logDepositor: mockLogDepositor,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_SetHeapPages_DecodeArgs_Success(t *testing.T) {
	call := setupCallSetHeapPages()

	buf := bytes.NewBuffer(someSetHeapPagesArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someSetHeapPagesArgs, call.Args())
}

func Test_Call_SetHeapPages_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionSetHeapPagesIndex}, defaultSetHeapPagesArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallSetHeapPages()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_SetHeapPages_Encode_With_Args(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionSetHeapPagesIndex}, someSetHeapPagesArgs.Bytes()...))
	buf := bytes.NewBuffer(someSetHeapPagesArgs.Bytes())

	call := setupCallSetHeapPages()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_SetHeapPages_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSetHeapPagesIndex}, defaultSetHeapPagesArgs.Bytes()...)

	call := setupCallSetHeapPages()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_SetHeapPages_ModuleIndex(t *testing.T) {
	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallSetHeapPages(tc, functionSetHeapPagesIndex, dbWeight, mockStorageHeapPages, mockLogDepositor)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_SetHeapPages_FunctionIndex(t *testing.T) {
	testCases := []sc.U8{
		0,
		functionSetHeapPagesIndex,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallSetHeapPages(moduleId, tc, dbWeight, mockStorageHeapPages, mockLogDepositor)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_SetHeapPages_BaseWeight(t *testing.T) {
	call := setupCallSetHeapPages()

	assert.Equal(t, callSetHeapPagesWeight(dbWeight), call.BaseWeight())
}

func Test_Call_SetHeapPages_WeighData(t *testing.T) {
	call := setupCallSetHeapPages()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_SetHeapPages_ClassifyDispatch(t *testing.T) {
	call := setupCallSetHeapPages()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_SetHeapPages_PaysFee(t *testing.T) {
	call := setupCallSetHeapPages()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_SetHeapPages_Dispatch_Success(t *testing.T) {
	call := setupCallSetHeapPages()
	call, err := call.DecodeArgs(bytes.NewBuffer(someSetHeapPagesArgs.Bytes()))
	assert.Nil(t, err)

	mockStorageHeapPages.On("Put", pages).Return()
	mockLogDepositor.On("DepositLog", digestItem).Return()

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	assert.Nil(t, dispatchErr)
	mockStorageHeapPages.AssertCalled(t, "Put", pages)
	mockLogDepositor.AssertCalled(t, "DepositLog", digestItem)
}

func setupCallSetHeapPages() primitives.Call {
	mockLogDepositor = new(mocks.SystemModule)
	return newCallSetHeapPages(moduleId, functionSetHeapPagesIndex, dbWeight, mockStorageHeapPages, mockLogDepositor)
}
