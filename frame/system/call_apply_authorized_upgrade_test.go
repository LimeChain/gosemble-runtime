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
	someApplyAuthorizedUpgradeArgs    = sc.NewVaryingData(codeBlob)
	defaultApplyAuthorizedUpgradeArgs = sc.NewVaryingData(sc.Sequence[sc.U8]{})
)

func Test_Call_ApplyAuthorizedUpgrade_New(t *testing.T) {
	call := setupCallApplyAuthorizedUpgrade()

	expected := callApplyAuthorizedUpgrade{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionApplyAuthorizedUpgradeIndex,
			Arguments:  defaultApplyAuthorizedUpgradeArgs,
		},
		dbWeight:     dbWeight,
		codeUpgrader: codeUpgrader,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_ApplyAuthorizedUpgrade_DecodeArgs_Success(t *testing.T) {
	call := setupCallApplyAuthorizedUpgrade()
	assert.Equal(t, defaultApplyAuthorizedUpgradeArgs, call.Args())

	buf := bytes.NewBuffer(someApplyAuthorizedUpgradeArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someApplyAuthorizedUpgradeArgs, call.Args())
}

func Test_Call_ApplyAuthorizedUpgrade_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionApplyAuthorizedUpgradeIndex}, defaultApplyAuthorizedUpgradeArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallApplyAuthorizedUpgrade()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_ApplyAuthorizedUpgrade_Encode_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionApplyAuthorizedUpgradeIndex}, someApplyAuthorizedUpgradeArgs.Bytes()...))
	buf := bytes.NewBuffer(someApplyAuthorizedUpgradeArgs.Bytes())

	call := setupCallApplyAuthorizedUpgrade()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_ApplyAuthorizedUpgrade_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionApplyAuthorizedUpgradeIndex}, defaultApplyAuthorizedUpgradeArgs.Bytes()...)

	call := setupCallApplyAuthorizedUpgrade()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_ApplyAuthorizedUpgrade_ModuleIndex(t *testing.T) {
	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallApplyAuthorizedUpgrade(tc, functionApplyAuthorizedUpgradeIndex, dbWeight, codeUpgrader)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_ApplyAuthorizedUpgrade_FunctionIndex(t *testing.T) {
	testCases := []sc.U8{
		7,
		8,
		9,
		functionApplyAuthorizedUpgradeIndex,
	}

	for _, tc := range testCases {
		call := newCallApplyAuthorizedUpgrade(moduleId, tc, dbWeight, codeUpgrader)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_ApplyAuthorizedUpgrade_BaseWeight(t *testing.T) {
	call := setupCallApplyAuthorizedUpgrade()

	assert.Equal(t, callApplyAuthorizedUpgradeWeight(dbWeight), call.BaseWeight())
}

func Test_Call_ApplyAuthorizedUpgrade_WeighData(t *testing.T) {
	call := setupCallApplyAuthorizedUpgrade()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_ApplyAuthorizedUpgrade_ClassifyDispatch(t *testing.T) {
	call := setupCallApplyAuthorizedUpgrade()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_ApplyAuthorizedUpgrade_PaysFee(t *testing.T) {
	call := setupCallApplyAuthorizedUpgrade()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_ApplyAuthorizedUpgrade_Dispatch(t *testing.T) {
	call := setupCallApplyAuthorizedUpgrade()
	call, err := call.DecodeArgs(bytes.NewBuffer(someApplyAuthorizedUpgradeArgs.Bytes()))
	assert.Nil(t, err)

	post := primitives.PostDispatchInfo{}
	codeUpgrader.On("DoApplyAuthorizeUpgrade", codeBlob).Return(post, nil)

	res, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	assert.Nil(t, dispatchErr)
	assert.Equal(t, post, res)
	codeUpgrader.AssertCalled(t, "DoApplyAuthorizeUpgrade", codeBlob)
}

func setupCallApplyAuthorizedUpgrade() primitives.Call {
	codeUpgrader = new(mocks.SystemModule)
	return newCallApplyAuthorizedUpgrade(moduleId, functionApplyAuthorizedUpgradeIndex, dbWeight, codeUpgrader)
}
