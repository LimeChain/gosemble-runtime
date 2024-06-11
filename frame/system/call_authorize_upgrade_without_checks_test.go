package system

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func Test_Call_AuthorizeUpgradeWithoutChecks_New(t *testing.T) {
	call := setupCallAuthorizeUpgradeWithoutChecks()

	expected := callAuthorizeUpgradeWithoutChecks{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionAuthorizeUpgradeWithoutChecksIndex,
			Arguments:  defaultAuthorizeUpgradeArgs,
		},
		dbWeight:     dbWeight,
		codeUpgrader: codeUpgrader,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_AuthorizeUpgradeWithoutChecks_Success(t *testing.T) {
	call := setupCallAuthorizeUpgradeWithoutChecks()
	assert.Equal(t, defaultAuthorizeUpgradeArgs, call.Args())

	buf := bytes.NewBuffer(someAuthorizeUpgradeArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someAuthorizeUpgradeArgs, call.Args())
}

func Test_Call_AuthorizeUpgradeWithoutChecks_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionAuthorizeUpgradeWithoutChecksIndex}, defaultAuthorizeUpgradeArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallAuthorizeUpgradeWithoutChecks()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_AuthorizeUpgradeWithoutChecks_Encode_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionAuthorizeUpgradeWithoutChecksIndex}, someAuthorizeUpgradeArgs.Bytes()...))
	buf := bytes.NewBuffer(someAuthorizeUpgradeArgs.Bytes())

	call := setupCallAuthorizeUpgradeWithoutChecks()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_AuthorizeUpgradeWithoutChecks_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionAuthorizeUpgradeWithoutChecksIndex}, defaultAuthorizeUpgradeArgs.Bytes()...)

	call := setupCallAuthorizeUpgradeWithoutChecks()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_AuthorizeUpgradeWithoutChecks_ModuleIndex(t *testing.T) {
	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallAuthorizeUpgradeWithoutChecks(tc, functionAuthorizeUpgradeWithoutChecksIndex, dbWeight, codeUpgrader)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_AuthorizeUpgradeWithoutChecks_FunctionIndex(t *testing.T) {
	testCases := []sc.U8{
		7,
		8,
		9,
		functionAuthorizeUpgradeWithoutChecksIndex,
	}

	for _, tc := range testCases {
		call := newCallAuthorizeUpgradeWithoutChecks(moduleId, tc, dbWeight, codeUpgrader)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_AuthorizeUpgradeWithoutChecks_BaseWeight(t *testing.T) {
	call := setupCallAuthorizeUpgradeWithoutChecks()

	assert.Equal(t, callAuthorizeUpgradeWithoutChecksWeight(dbWeight), call.BaseWeight())
}

func Test_Call_AuthorizeUpgradeWithoutChecks_WeighData(t *testing.T) {
	call := setupCallAuthorizeUpgradeWithoutChecks()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_AuthorizeUpgradeWithoutChecks_ClassifyDispatch(t *testing.T) {
	call := setupCallAuthorizeUpgradeWithoutChecks()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_AuthorizeUpgradeWithoutChecks_PaysFee(t *testing.T) {
	call := setupCallAuthorizeUpgradeWithoutChecks()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_AuthorizeUpgradeWithoutChecks_Dispatch(t *testing.T) {
	call := setupCallAuthorizeUpgradeWithoutChecks()
	call, err := call.DecodeArgs(bytes.NewBuffer(someAuthorizeUpgradeArgs.Bytes()))
	assert.Nil(t, err)

	codeUpgrader.On("DoAuthorizeUpgrade", codeHash, sc.Bool(false)).Return()

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	assert.Nil(t, dispatchErr)
	codeUpgrader.AssertCalled(t, "DoAuthorizeUpgrade", codeHash, sc.Bool(false))
}

func Test_Call_AuthorizeUpgradeWithoutChecks_Dispatch_BadOrigin(t *testing.T) {
	call := setupCallAuthorizeUpgradeWithoutChecks()
	call, err := call.DecodeArgs(bytes.NewBuffer(someAuthorizeUpgradeArgs.Bytes()))
	assert.Nil(t, err)

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginNone(), call.Args())

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), dispatchErr)
	codeUpgrader.AssertNotCalled(t, "DoAuthorizeUpgrade", codeHash, sc.Bool(false))
}

func setupCallAuthorizeUpgradeWithoutChecks() primitives.Call {
	codeUpgrader = new(mocks.SystemModule)
	return newCallAuthorizeUpgradeWithoutChecks(moduleId, functionAuthorizeUpgradeWithoutChecksIndex, dbWeight, codeUpgrader)
}
