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
	codeHash, _ = primitives.NewH256(sc.BytesToFixedSequenceU8(hashBytes)...)

	someAuthorizeUpgradeArgs    = sc.NewVaryingData(codeHash)
	defaultAuthorizeUpgradeArgs = sc.NewVaryingData(primitives.H256{})
)

var (
	codeUpgrader *mocks.SystemModule
)

func Test_Call_AuthorizeUpgrade_New(t *testing.T) {
	call := setupCallAuthorizeUpgrade()

	expected := callAuthorizeUpgrade{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionAuthorizeUpgradeIndex,
			Arguments:  defaultAuthorizeUpgradeArgs,
		},
		dbWeight:     dbWeight,
		codeUpgrader: codeUpgrader,
	}

	assert.Equal(t, expected, call)
}

func Test_Call_AuthorizeUpgrade_DecodeArgs_Success(t *testing.T) {
	call := setupCallAuthorizeUpgrade()
	assert.Equal(t, defaultAuthorizeUpgradeArgs, call.Args())

	buf := bytes.NewBuffer(someAuthorizeUpgradeArgs.Bytes())
	call, err := call.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, someAuthorizeUpgradeArgs, call.Args())
}

func Test_Call_AuthorizeUpgrade_Encode(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionAuthorizeUpgradeIndex}, defaultAuthorizeUpgradeArgs.Bytes()...))
	buf := &bytes.Buffer{}

	call := setupCallAuthorizeUpgrade()
	err := call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_AuthorizeUpgrade_Encode_WithArgs(t *testing.T) {
	expectedBuf := bytes.NewBuffer(append([]byte{moduleId, functionAuthorizeUpgradeIndex}, someAuthorizeUpgradeArgs.Bytes()...))
	buf := bytes.NewBuffer(someAuthorizeUpgradeArgs.Bytes())

	call := setupCallAuthorizeUpgrade()
	call, err := call.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	err = call.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, expectedBuf.Bytes(), buf.Bytes())
}

func Test_Call_AuthorizeUpgrade_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionAuthorizeUpgradeIndex}, defaultAuthorizeUpgradeArgs.Bytes()...)

	call := setupCallAuthorizeUpgrade()

	assert.Equal(t, expected, call.Bytes())
}

func Test_Call_AuthorizeUpgrade_ModuleIndex(t *testing.T) {
	testCases := []sc.U8{
		moduleId,
		1,
		2,
		3,
	}

	for _, tc := range testCases {
		call := newCallAuthorizeUpgrade(tc, functionAuthorizeUpgradeIndex, dbWeight, codeUpgrader)

		assert.Equal(t, tc, call.ModuleIndex())
	}
}

func Test_Call_AuthorizeUpgrade_FunctionIndex(t *testing.T) {
	testCases := []sc.U8{
		6,
		7,
		8,
		functionAuthorizeUpgradeIndex,
	}

	for _, tc := range testCases {
		call := newCallAuthorizeUpgrade(moduleId, tc, dbWeight, codeUpgrader)

		assert.Equal(t, tc, call.FunctionIndex())
	}
}

func Test_Call_AuthorizeUpgrade_BaseWeight(t *testing.T) {
	call := setupCallAuthorizeUpgrade()

	assert.Equal(t, callAuthorizeUpgradeWeight(dbWeight), call.BaseWeight())
}

func Test_Call_AuthorizeUpgrade_WeighData(t *testing.T) {
	call := setupCallAuthorizeUpgrade()

	assert.Equal(t, primitives.WeightFromParts(567, 0), call.WeighData(baseWeight))
}

func Test_Call_AuthorizeUpgrade_ClassifyDispatch(t *testing.T) {
	call := setupCallAuthorizeUpgrade()

	assert.Equal(t, primitives.NewDispatchClassOperational(), call.ClassifyDispatch(baseWeight))
}

func Test_Call_AuthorizeUpgrade_PaysFee(t *testing.T) {
	call := setupCallAuthorizeUpgrade()

	assert.Equal(t, primitives.PaysNo, call.PaysFee(baseWeight))
}

func Test_Call_AuthorizeUpgrade_Dispatch(t *testing.T) {
	call := setupCallAuthorizeUpgrade()
	call, err := call.DecodeArgs(bytes.NewBuffer(someAuthorizeUpgradeArgs.Bytes()))
	assert.Nil(t, err)

	codeUpgrader.On("DoAuthorizeUpgrade", codeHash, sc.Bool(true)).Return()

	_, dispatchErr := call.Dispatch(primitives.NewRawOriginRoot(), call.Args())

	assert.Nil(t, dispatchErr)
	codeUpgrader.AssertCalled(t, "DoAuthorizeUpgrade", codeHash, sc.Bool(true))
}

func setupCallAuthorizeUpgrade() primitives.Call {
	codeUpgrader = new(mocks.SystemModule)
	return newCallAuthorizeUpgrade(moduleId, functionAuthorizeUpgradeIndex, dbWeight, codeUpgrader)
}
