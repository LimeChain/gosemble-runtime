package extensions

import (
	"bytes"
	"errors"
	"math"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	invalidTransactionStale  = primitives.NewTransactionValidityError(primitives.NewInvalidTransactionStale())
	invalidTransactionFuture = primitives.NewTransactionValidityError(primitives.NewInvalidTransactionFuture())
)

var (
	oneAccountId = constants.OneAccountId
)

func Test_CheckNonce_Encode(t *testing.T) {
	nonce := sc.U32(1)
	buffer := &bytes.Buffer{}

	target := setupCheckNonce()
	target.nonce = nonce

	err := target.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, sc.ToCompact(nonce).Bytes(), buffer.Bytes())
}

func Test_CheckNonce_Empty(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := setupCheckNonce().Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, sc.ToCompact(sc.U32(0)).Bytes(), buffer.Bytes())
}

func Test_CheckNonce_Decode(t *testing.T) {
	nonce := sc.U32(1)
	buffer := bytes.NewBuffer(sc.ToCompact(nonce).Bytes())

	target := setupCheckNonce()

	target.Decode(buffer)

	assert.Equal(t, nonce, target.nonce)
}

func Test_CheckNonce_Bytes(t *testing.T) {
	nonce := sc.U32(1)
	target := setupCheckNonce()
	target.nonce = nonce

	result := target.Bytes()

	assert.Equal(t, sc.ToCompact(nonce).Bytes(), result)
}

func Test_CheckNonce_DeepCopy(t *testing.T) {
	nonce := sc.U32(1)
	target := setupCheckNonce()
	target.nonce = nonce

	result := target.DeepCopy()

	assert.Equal(t, &target, result)

	target.nonce = 120
	assert.NotEqual(t, &target, result)
}

func Test_CheckNonce_AdditionalSigned(t *testing.T) {
	target := setupCheckNonce()

	result, err := target.AdditionalSigned()

	assert.Nil(t, err)
	assert.Equal(t, primitives.AdditionalSigned{}, result)
}

func Test_CheckNonce_Validate_WithRequires_Success(t *testing.T) {
	nonce := sc.U32(1)
	accountInfo := primitives.AccountInfo{
		Nonce: 0,
	}
	expect := primitives.ValidTransaction{
		Priority: 0,
		Provides: sc.Sequence[sc.Sequence[sc.U8]]{
			sc.BytesToSequenceU8(append(oneAccountId.Bytes(), sc.ToCompact(nonce).Bytes()...)),
		},
		Requires: sc.Sequence[sc.Sequence[sc.U8]]{
			sc.BytesToSequenceU8(append(oneAccountId.Bytes(), sc.ToCompact(nonce-1).Bytes()...)),
		},
		Longevity: math.MaxUint64,
		Propagate: true,
	}

	target := setupCheckNonce()
	target.nonce = nonce

	mockModule.On("StorageAccount", oneAccountId).Return(accountInfo, nil)

	result, err := target.Validate(oneAccountId, nil, nil, sc.Compact{})

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
	mockModule.AssertCalled(t, "StorageAccount", oneAccountId)
}

func Test_CheckNonce_Validate_NoRequires_Success(t *testing.T) {
	nonce := sc.U32(1)
	accountInfo := primitives.AccountInfo{
		Nonce: nonce,
	}
	expect := primitives.ValidTransaction{
		Priority: 0,
		Provides: sc.Sequence[sc.Sequence[sc.U8]]{
			sc.BytesToSequenceU8(append(oneAccountId.Bytes(), sc.ToCompact(nonce).Bytes()...)),
		},
		Requires:  sc.Sequence[sc.Sequence[sc.U8]]{},
		Longevity: math.MaxUint64,
		Propagate: true,
	}

	target := setupCheckNonce()
	target.nonce = nonce

	mockModule.On("StorageAccount", oneAccountId).Return(accountInfo, nil)

	result, err := target.Validate(oneAccountId, nil, nil, sc.Compact{})

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
	mockModule.AssertCalled(t, "StorageAccount", oneAccountId)
}

func Test_CheckNonce_Validate_Fails(t *testing.T) {
	nonce := sc.U32(0)
	accountInfo := primitives.AccountInfo{
		Nonce: 1,
	}

	target := setupCheckNonce()
	target.nonce = nonce

	mockModule.On("StorageAccount", oneAccountId).Return(accountInfo, nil)

	result, err := target.Validate(oneAccountId, nil, nil, sc.Compact{})

	assert.Equal(t, invalidTransactionStale, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
	mockModule.AssertCalled(t, "StorageAccount", oneAccountId)
}

func Test_CheckNonce_Validate_Fails_StorageAccountError(t *testing.T) {
	nonce := sc.U32(0)
	accountInfo := primitives.AccountInfo{
		Nonce: 1,
	}

	target := setupCheckNonce()
	target.nonce = nonce

	expectedErr := errors.New("error")
	mockModule.On("StorageAccount", oneAccountId).Return(accountInfo, expectedErr)

	_, err := target.Validate(oneAccountId, nil, nil, sc.Compact{})

	assert.Equal(t, expectedErr, err)
	mockModule.AssertCalled(t, "StorageAccount", oneAccountId)
}

func Test_CheckNonce_ValidateUnsigned(t *testing.T) {
	target := setupCheckNonce()

	result, err := target.ValidateUnsigned(nil, nil, sc.Compact{})

	assert.Nil(t, err)
	assert.Equal(t, primitives.DefaultValidTransaction(), result)
}

func Test_CheckNonce_PreDispatch_Success(t *testing.T) {
	nonce := sc.U32(1)
	accountInfo := primitives.AccountInfo{
		Nonce: 1,
	}
	expectAccountInfo := primitives.AccountInfo{
		Nonce: accountInfo.Nonce + 1,
	}

	target := setupCheckNonce()
	target.nonce = nonce

	mockModule.On("StorageAccount", oneAccountId).Return(accountInfo, nil)
	mockModule.On("StorageAccountSet", oneAccountId, expectAccountInfo)

	result, err := target.PreDispatch(oneAccountId, nil, nil, sc.Compact{})

	assert.Nil(t, err)
	assert.Equal(t, primitives.Pre{}, result)

	mockModule.AssertCalled(t, "StorageAccount", oneAccountId)
	mockModule.AssertCalled(t, "StorageAccountSet", oneAccountId, expectAccountInfo)
}

func Test_CheckNonce_PreDispatch_Fails_Stale(t *testing.T) {
	nonce := sc.U32(0)
	accountInfo := primitives.AccountInfo{
		Nonce: 1,
	}

	target := setupCheckNonce()
	target.nonce = nonce

	mockModule.On("StorageAccount", oneAccountId).Return(accountInfo, nil)

	result, err := target.PreDispatch(oneAccountId, nil, nil, sc.Compact{})

	assert.Equal(t, invalidTransactionStale, err)
	assert.Equal(t, primitives.Pre{}, result)

	mockModule.AssertCalled(t, "StorageAccount", oneAccountId)
	mockModule.AssertNotCalled(t, "StorageAccountSet", oneAccountId, mock.Anything)
}

func Test_CheckNonce_PreDispatch_Fails_Future(t *testing.T) {
	nonce := sc.U32(2)
	accountInfo := primitives.AccountInfo{
		Nonce: 1,
	}

	target := setupCheckNonce()
	target.nonce = nonce

	mockModule.On("StorageAccount", oneAccountId).Return(accountInfo, nil)

	result, err := target.PreDispatch(oneAccountId, nil, nil, sc.Compact{})

	assert.Equal(t, invalidTransactionFuture, err)
	assert.Equal(t, primitives.Pre{}, result)

	mockModule.AssertCalled(t, "StorageAccount", oneAccountId)
	mockModule.AssertNotCalled(t, "StorageAccountSet", oneAccountId, mock.Anything)
}

func Test_CheckNonce_PreDispatch_Fails_StorageAccountError(t *testing.T) {
	nonce := sc.U32(2)
	accountInfo := primitives.AccountInfo{
		Nonce: 1,
	}

	target := setupCheckNonce()
	target.nonce = nonce

	expectedErr := errors.New("error")
	mockModule.On("StorageAccount", oneAccountId).Return(accountInfo, expectedErr)

	_, err := target.PreDispatch(oneAccountId, nil, nil, sc.Compact{})

	assert.Equal(t, expectedErr, err)

	mockModule.AssertCalled(t, "StorageAccount", oneAccountId)
}

func Test_CheckNonce_PreDispatchUnsigned(t *testing.T) {
	target := setupCheckNonce()

	err := target.PreDispatchUnsigned(nil, nil, sc.Compact{})

	assert.Nil(t, err)
}

func Test_CheckNonce_PostDispatch(t *testing.T) {
	target := setupCheckNonce()

	err := target.PostDispatch(sc.NewOption[primitives.Pre](nil), nil, nil, sc.Compact{}, nil)

	assert.Nil(t, err)
}

func Test_CheckNonce_ModulePath(t *testing.T) {
	target := setupCheckNonce()

	expectedModulePath := "frame_system"
	actualModulePath := target.ModulePath()

	assert.Equal(t, expectedModulePath, actualModulePath)
}

func setupCheckNonce() CheckNonce {
	mockModule = new(mocks.SystemModule)
	extension, ok := NewCheckNonce(mockModule).(*CheckNonce)
	if !ok {
		panic("invalid type assert for *CheckNonce")
	}
	return *extension
}
