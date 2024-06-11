package system

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	remarkMsg         = sc.BytesToSequenceU8([]byte("remark messsage"))
	someRemarkArgs    = sc.NewVaryingData(remarkMsg)
	defaultRemarkArgs = sc.NewVaryingData(sc.Sequence[sc.U8]{})
)

func Test_Call_Remark_New(t *testing.T) {
	expected := callRemark{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionRemarkIndex,
			Arguments:  defaultRemarkArgs,
		},
		dbWeight: dbWeight,
	}

	call := setupCallRemark()

	assert.Equal(t, expected, call)
}

func Test_Call_Remark_DecodeArgs_Success(t *testing.T) {
	buf := bytes.NewBuffer(someRemarkArgs.Bytes())

	call := setupCallRemark()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	assert.Equal(t, someRemarkArgs, call.Args())
}

func Test_Call_Remark_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionRemarkIndex}, defaultRemarkArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallRemark()
	err := call.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuf, buf)
}

func Test_Call_Remark_EncodeWithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionRemarkIndex}, someRemarkArgs.Bytes()...))
	buf := bytes.NewBuffer(someRemarkArgs.Bytes())

	call := setupCallRemark()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuf, buf)
}

func Test_Call_Remark_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionRemarkIndex}, defaultRemarkArgs.Bytes()...)

	call := setupCallRemark()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_Remark_ModuleIndex(t *testing.T) {
	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallRemark(tc, functionRemarkIndex, dbWeight)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_Remark_FunctionIndex(t *testing.T) {
	testCases := []sc.U8{
		functionRemarkIndex,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallRemark(moduleId, tc, dbWeight)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_Remark_BaseWeight(t *testing.T) {
	call := setupCallRemark()

	assert.Equal(t, callRemarkWeight(dbWeight, 0), call.BaseWeight())
}

func Test_Call_Remark_WeighData(t *testing.T) {
	call := setupCallRemark()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_Remark_ClassifyDispatch(t *testing.T) {
	call := setupCallRemark()

	assert.Equal(t, primitives.NewDispatchClassNormal(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_Remark_PaysFee(t *testing.T) {
	call := setupCallRemark()

	assert.Equal(t, primitives.PaysYes, call.PaysFee(baseWeight))
}

func Test_Call_Remark_Dispatch_AnyOrigin_Success(t *testing.T) {
	call := setupCallRemark()

	for _, origin := range []primitives.RuntimeOrigin{
		primitives.NewRawOriginNone(),
		primitives.NewRawOriginRoot(),
		primitives.NewRawOriginSigned(primitives.NewAccountIdFromAddress32(addr)),
	} {
		_, dispatchErr := call.Dispatch(origin, nil)

		assert.Nil(t, dispatchErr)
	}
}

func setupCallRemark() primitives.Call {
	return newCallRemark(moduleId, functionRemarkIndex, dbWeight)
}
