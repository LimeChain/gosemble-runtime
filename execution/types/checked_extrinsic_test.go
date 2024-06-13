package types

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	signerOption = sc.NewOption[types.AccountId](constants.ZeroAccountId)
	emptySigner  = sc.NewOption[types.AccountId](nil)

	txSource     = types.NewTransactionSourceExternal()
	dispatchInfo = &types.DispatchInfo{
		Weight:  types.WeightFromParts(4, 5),
		Class:   types.NewDispatchClassMandatory(),
		PaysFee: types.PaysNo,
	}
	length             = sc.ToCompact(5)
	postDispatchInfoOk = types.PostDispatchInfo{
		ActualWeight: sc.NewOption[types.Weight](types.WeightFromParts(2, 3)),
		PaysFee:      types.PaysYes,
	}
	postDispatchInfoErr = types.PostDispatchInfo{
		ActualWeight: sc.NewOption[types.Weight](nil),
		PaysFee:      types.PaysNo,
	}
	errPostDispatch = types.NewDispatchErrorCorruption()
	pre             = sc.Sequence[types.Pre]{sc.NewVaryingData(sc.U32(1)), sc.NewVaryingData(sc.U32(7))}
	optionPre       = sc.NewOption[sc.Sequence[types.Pre]](pre)
	emptyOptionPre  = sc.NewOption[sc.Sequence[types.Pre]](nil)
)

var (
	expectedInvalidTransactionStaleErr               = types.NewTransactionValidityError(types.NewInvalidTransactionStale())
	expectedInvalidTransactionPaymentErr             = types.NewTransactionValidityError(types.NewInvalidTransactionPayment())
	expectedUnknownTransactionNoUnsignedValidatorErr = types.NewTransactionValidityError(types.NewUnknownTransactionNoUnsignedValidator())
)

var (
	mockStorage           *mocks.IoStorage
	mockTransactionBroker *mocks.IoTransactionBroker
	mockTransactional     *mocks.IoTransactional[types.PostDispatchInfo]
	mockUnsignedValidator *mocks.UnsignedValidator

	mockWithStorageLayer = mock.AnythingOfType("func() (types.PostDispatchInfo, error)")
)

func Test_CheckedExtrinsic_Function(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	assert.Equal(t, mockCall, target.Function())
}

func Test_CheckedExtrinsic_Apply_Signed_Success(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	mockSignedExtra.
		On("PreDispatch", signerOption.Value, mockCall, dispatchInfo, length).
		Return(pre, nil)
	mockTransactional.On("WithStorageLayer", mockWithStorageLayer).Return(postDispatchInfoOk, nil)
	mockSignedExtra.
		On("PostDispatch", optionPre, dispatchInfo, &postDispatchInfoOk, length, nil).
		Return(nil)

	result, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)
	assert.Equal(t, nil, err)
	assert.Equal(t, postDispatchInfoOk, result)
	mockSignedExtra.
		AssertCalled(t, "PreDispatch", signerOption.Value, mockCall, dispatchInfo, length)
	mockTransactional.AssertCalled(t, "WithStorageLayer", mockWithStorageLayer)
	mockSignedExtra.
		AssertCalled(t, "PostDispatch", optionPre, dispatchInfo, &postDispatchInfoOk, length, nil)
}

func Test_CheckedExtrinsic_Apply_Signed_PreDispatch_Fails(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	mockSignedExtra.
		On("PreDispatch", signerOption.Value, mockCall, dispatchInfo, length).
		Return(pre, expectedInvalidTransactionStaleErr)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, expectedInvalidTransactionStaleErr, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatch", signerOption.Value, mockCall, dispatchInfo, length)
	mockTransactional.AssertNotCalled(t, "WithStorageLayer", mock.Anything)
}

func Test_CheckedExtrinsic_Apply_Signed_WithStorageLayerErr(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	mockSignedExtra.
		On("PreDispatch", signerOption.Value, mockCall, dispatchInfo, length).
		Return(pre, nil)
	mockTransactional.On("WithStorageLayer", mockWithStorageLayer).Return(postDispatchInfoErr, errPostDispatch)
	mockSignedExtra.
		On("PostDispatch", optionPre, dispatchInfo, &postDispatchInfoErr, length, errPostDispatch).
		Return(nil)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, errPostDispatch, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatch", signerOption.Value, mockCall, dispatchInfo, length)
	mockTransactional.AssertCalled(t, "WithStorageLayer", mockWithStorageLayer)
	mockSignedExtra.
		AssertCalled(t, "PostDispatch", optionPre, dispatchInfo, &postDispatchInfoErr, length, errPostDispatch)
}

func Test_CheckedExtrinsic_Apply_Signed_WithStorageLayerErr_PostDispatchErr(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	mockSignedExtra.
		On("PreDispatch", signerOption.Value, mockCall, dispatchInfo, length).
		Return(pre, nil)
	mockTransactional.On("WithStorageLayer", mockWithStorageLayer).Return(postDispatchInfoErr, errPostDispatch)
	mockSignedExtra.
		On("PostDispatch", optionPre, dispatchInfo, &postDispatchInfoErr, length, errPostDispatch).
		Return(expectedInvalidTransactionStaleErr)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, expectedInvalidTransactionStaleErr, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatch", signerOption.Value, mockCall, dispatchInfo, length)
	mockTransactional.AssertCalled(t, "WithStorageLayer", mockWithStorageLayer)
	mockSignedExtra.
		AssertCalled(t, "PostDispatch", optionPre, dispatchInfo, &postDispatchInfoErr, length, errPostDispatch)
}

func Test_CheckedExtrinsic_Apply_Signed_PostDispatchFails(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	mockSignedExtra.
		On("PreDispatch", signerOption.Value, mockCall, dispatchInfo, length).
		Return(pre, nil)
	mockTransactional.On("WithStorageLayer", mockWithStorageLayer).Return(postDispatchInfoOk, nil)
	mockSignedExtra.
		On("PostDispatch", optionPre, dispatchInfo, &postDispatchInfoOk, length, nil).
		Return(expectedInvalidTransactionStaleErr)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, expectedInvalidTransactionStaleErr, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatch", signerOption.Value, mockCall, dispatchInfo, length)
	mockTransactional.AssertCalled(t, "WithStorageLayer", mockWithStorageLayer)
	mockSignedExtra.
		AssertCalled(t, "PostDispatch", optionPre, dispatchInfo, &postDispatchInfoOk, length, nil)
}

func Test_CheckedExtrinsic_Apply_Unsigned_Success(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	mockSignedExtra.
		On("PreDispatchUnsigned", mockCall, dispatchInfo, length).
		Return(nil)
	mockUnsignedValidator.On("PreDispatch", mockCall).Return(sc.Empty{}, nil)
	mockTransactional.On("WithStorageLayer", mockWithStorageLayer).Return(postDispatchInfoOk, nil)
	mockSignedExtra.
		On("PostDispatch", emptyOptionPre, dispatchInfo, &postDispatchInfoOk, length, nil).
		Return(nil)

	result, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, nil, err)
	assert.Equal(t, postDispatchInfoOk, result)
	mockSignedExtra.
		AssertCalled(t, "PreDispatchUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.AssertCalled(t, "PreDispatch", mockCall)
	mockTransactional.AssertCalled(t, "WithStorageLayer", mockWithStorageLayer)
	mockSignedExtra.
		AssertCalled(t, "PostDispatch", emptyOptionPre, dispatchInfo, &postDispatchInfoOk, length, nil)
}

func Test_CheckedExtrinsic_Apply_Unsigned_PreDispatchUnsigned_Fails(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	mockSignedExtra.
		On("PreDispatchUnsigned", mockCall, dispatchInfo, length).
		Return(expectedInvalidTransactionStaleErr)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, expectedInvalidTransactionStaleErr, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatchUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.AssertNotCalled(t, "PreDispatch", mock.Anything)
	mockTransactional.AssertNotCalled(t, "WithStorageLayer", mock.Anything)
}

func Test_CheckedExtrinsic_Apply_Unsigned_PreDispatch_Fails(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	mockSignedExtra.
		On("PreDispatchUnsigned", mockCall, dispatchInfo, length).
		Return(nil)
	mockUnsignedValidator.On("PreDispatch", mockCall).Return(sc.Empty{}, expectedInvalidTransactionStaleErr)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, expectedInvalidTransactionStaleErr, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatchUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.AssertCalled(t, "PreDispatch", mockCall)
	mockTransactional.AssertNotCalled(t, "WithStorageLayer", mock.Anything)
}

func Test_CheckedExtrinsic_Apply_Unsigned_WithStorageLayerErr(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	mockSignedExtra.
		On("PreDispatchUnsigned", mockCall, dispatchInfo, length).
		Return(nil)
	mockUnsignedValidator.On("PreDispatch", mockCall).Return(sc.Empty{}, nil)
	mockTransactional.On("WithStorageLayer", mockWithStorageLayer).Return(postDispatchInfoErr, errPostDispatch)
	mockSignedExtra.
		On("PostDispatch", emptyOptionPre, dispatchInfo, &postDispatchInfoErr, length, errPostDispatch).
		Return(nil)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, errPostDispatch, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatchUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.AssertCalled(t, "PreDispatch", mockCall)
	mockTransactional.AssertCalled(t, "WithStorageLayer", mockWithStorageLayer)
	mockSignedExtra.
		AssertCalled(t, "PostDispatch", emptyOptionPre, dispatchInfo, &postDispatchInfoErr, length, errPostDispatch)
}

func Test_CheckedExtrinsic_Apply_Unsigned_WithStorageLayerErr_PostDispatch_Fails(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	mockSignedExtra.
		On("PreDispatchUnsigned", mockCall, dispatchInfo, length).
		Return(nil)
	mockUnsignedValidator.On("PreDispatch", mockCall).Return(sc.Empty{}, nil)
	mockTransactional.On("WithStorageLayer", mockWithStorageLayer).Return(postDispatchInfoErr, errPostDispatch)
	mockSignedExtra.
		On("PostDispatch", emptyOptionPre, dispatchInfo, &postDispatchInfoErr, length, errPostDispatch).
		Return(expectedInvalidTransactionStaleErr)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, expectedInvalidTransactionStaleErr, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatchUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.AssertCalled(t, "PreDispatch", mockCall)
	mockTransactional.AssertCalled(t, "WithStorageLayer", mockWithStorageLayer)
	mockSignedExtra.
		AssertCalled(t, "PostDispatch", emptyOptionPre, dispatchInfo, &postDispatchInfoErr, length, errPostDispatch)
}

func Test_CheckedExtrinsic_Apply_Unsigned_PostDispatch_Fails(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	mockSignedExtra.
		On("PreDispatchUnsigned", mockCall, dispatchInfo, length).
		Return(nil)
	mockUnsignedValidator.On("PreDispatch", mockCall).Return(sc.Empty{}, nil)
	mockTransactional.On("WithStorageLayer", mockWithStorageLayer).Return(postDispatchInfoOk, nil)
	mockSignedExtra.
		On("PostDispatch", emptyOptionPre, dispatchInfo, &postDispatchInfoOk, length, nil).
		Return(expectedInvalidTransactionStaleErr)

	_, err := target.Apply(mockUnsignedValidator, dispatchInfo, length)

	assert.Equal(t, expectedInvalidTransactionStaleErr, err)
	mockSignedExtra.
		AssertCalled(t, "PreDispatchUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.AssertCalled(t, "PreDispatch", mockCall)
	mockTransactional.AssertCalled(t, "WithStorageLayer", mockWithStorageLayer)
	mockSignedExtra.
		AssertCalled(t, "PostDispatch", emptyOptionPre, dispatchInfo, &postDispatchInfoOk, length, nil)
}

func Test_CheckedExtrinsic_Validate_Signed_Success(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	mockSignedExtra.
		On("Validate", signerOption.Value, mockCall, dispatchInfo, length).
		Return(types.DefaultValidTransaction(), nil)

	result, err := target.Validate(mockUnsignedValidator, txSource, dispatchInfo, length)

	assert.Nil(t, err)
	assert.Equal(t, types.DefaultValidTransaction(), result)
	mockSignedExtra.
		AssertCalled(t, "Validate", signerOption.Value, mockCall, dispatchInfo, length)
}

func Test_CheckedExtrinsic_Validate_Unsigned_Success(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	expect := types.DefaultValidTransaction().CombineWith(types.DefaultValidTransaction())

	mockSignedExtra.
		On("ValidateUnsigned", mockCall, dispatchInfo, length).
		Return(types.DefaultValidTransaction(), nil)
	mockUnsignedValidator.
		On("ValidateUnsigned", txSource, mockCall).
		Return(types.DefaultValidTransaction(), nil)

	result, err := target.Validate(mockUnsignedValidator, txSource, dispatchInfo, length)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
	mockSignedExtra.
		AssertCalled(t, "ValidateUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.
		AssertCalled(t, "ValidateUnsigned", txSource, mockCall)
}

func Test_CheckedExtrinsic_Validate_Unsigned_ValidateUnsigned_Fails(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	mockSignedExtra.
		On("ValidateUnsigned", mockCall, dispatchInfo, length).
		Return(types.ValidTransaction{}, expectedInvalidTransactionPaymentErr)

	result, err := target.Validate(mockUnsignedValidator, txSource, dispatchInfo, length)

	assert.Equal(t, expectedInvalidTransactionPaymentErr, err)
	assert.Equal(t, types.ValidTransaction{}, result)
	mockSignedExtra.
		AssertCalled(t, "ValidateUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.
		AssertNotCalled(t, "ValidateUnsigned", mock.Anything, mock.Anything)
}

func Test_CheckedExtrinsic_Validate_Unsigned_ValidatorValidateUnsigned_Fails(t *testing.T) {
	target := setupCheckedExtrinsic(emptySigner)

	mockSignedExtra.
		On("ValidateUnsigned", mockCall, dispatchInfo, length).
		Return(types.DefaultValidTransaction(), nil)
	mockUnsignedValidator.
		On("ValidateUnsigned", txSource, mockCall).
		Return(types.ValidTransaction{}, expectedUnknownTransactionNoUnsignedValidatorErr)

	result, err := target.Validate(mockUnsignedValidator, txSource, dispatchInfo, length)

	assert.Equal(t, expectedUnknownTransactionNoUnsignedValidatorErr, err)
	assert.Equal(t, types.ValidTransaction{}, result)
	mockSignedExtra.
		AssertCalled(t, "ValidateUnsigned", mockCall, dispatchInfo, length)
	mockUnsignedValidator.
		AssertCalled(t, "ValidateUnsigned", txSource, mockCall)
}

func Test_CheckedExtrinsic_dispatch_Success(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	args := sc.NewVaryingData(sc.U32(1))

	mockCall.On("Args").Return(args)
	mockCall.On("Dispatch", types.RawOriginFrom(signerOption), args).Return(postDispatchInfoOk, nil)

	res, err := target.dispatch(signerOption)

	assert.Nil(t, err)
	assert.Equal(t, postDispatchInfoOk, res)
	mockCall.AssertCalled(t, "Args")
	mockCall.AssertCalled(t, "Dispatch", types.RawOriginFrom(signerOption), args)
}

func Test_CheckedExtrinsic_dispatch_Fails(t *testing.T) {
	target := setupCheckedExtrinsic(signerOption)

	args := sc.NewVaryingData(sc.U32(1))

	mockCall.On("Args").Return(args)
	mockCall.On("Dispatch", types.RawOriginFrom(signerOption), args).Return(postDispatchInfoErr, errPostDispatch)

	res, err := target.dispatch(signerOption)

	assert.Equal(t, postDispatchInfoErr, res)
	assert.Equal(t, errPostDispatch, err)
	mockCall.AssertCalled(t, "Args")
	mockCall.AssertCalled(t, "Dispatch", types.RawOriginFrom(signerOption), args)
}

func setupCheckedExtrinsic(signer sc.Option[types.AccountId]) checkedExtrinsic {
	mockStorage = new(mocks.IoStorage)
	mockCall = new(mocks.Call)
	mockSignedExtra = new(mocks.SignedExtra)
	mockTransactional = new(mocks.IoTransactional[types.PostDispatchInfo])
	mockUnsignedValidator = new(mocks.UnsignedValidator)

	target := NewCheckedExtrinsic(signer, mockCall, mockSignedExtra, mockStorage, mockTransactionBroker, logger).(checkedExtrinsic)
	target.transactional = mockTransactional

	return target
}
