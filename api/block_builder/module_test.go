package blockbuilder

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	dataPtr    = int32(0)
	dataLen    = int32(1)
	ptrAndSize = int64(5)

	bUxt     = []byte("unchecked_extrinsic")
	uxt      = new(mocks.UncheckedExtrinsic)
	errPanic = errors.New("panic")
)

var (
	mockExecutive        *mocks.Executive
	mockRuntimeDecoder   *mocks.RuntimeDecoder
	mockRuntimeExtrinsic *mocks.RuntimeExtrinsic
	mockMemoryUtils      *mocks.MemoryTranslator
	mdGenerator          = primitives.NewMetadataTypeGenerator()
)

func Test_Module_Name(t *testing.T) {
	target := setup()

	result := target.Name()

	assert.Equal(t, ApiModuleName, result)
}

func Test_Module_Item(t *testing.T) {
	target := setup()

	hexName := common.MustBlake2b8([]byte(ApiModuleName))
	expect := primitives.NewApiItem(hexName, apiVersion)

	result := target.Item()

	assert.Equal(t, expect, result)
}

func Test_Module_ApplyExtrinsic_NilDispatchOutput(t *testing.T) {
	target := setup()

	bufferUxt := bytes.NewBuffer(bUxt)
	outcome, err := primitives.NewDispatchOutcome(sc.Empty{})
	assert.Nil(t, err)
	applyExtrinsicResultOutcome, err := primitives.NewApplyExtrinsicResult(outcome)
	assert.Nil(t, err)
	bExtrinsicResult := applyExtrinsicResultOutcome.Bytes()

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bUxt)
	mockRuntimeDecoder.On("DecodeUncheckedExtrinsic", bufferUxt).Return(uxt, nil)
	mockExecutive.On("ApplyExtrinsic", uxt).Return(nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", bExtrinsicResult).Return(ptrAndSize)

	result := target.ApplyExtrinsic(dataPtr, dataLen)

	assert.Equal(t, ptrAndSize, result)
	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockExecutive.AssertCalled(t, "ApplyExtrinsic", uxt)
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", bExtrinsicResult)
}

func Test_Module_ApplyExtrinsic_TransactionValidityError(t *testing.T) {
	target := setup()

	bufferUxt := bytes.NewBuffer(bUxt)
	validityError, ok := primitives.NewTransactionValidityError(primitives.NewInvalidTransactionStale()).(primitives.TransactionValidityError)
	assert.True(t, ok)
	applyExtrinsicResultValidityErr, err := primitives.NewApplyExtrinsicResult(validityError)
	assert.Nil(t, err)
	bExtrinsicResult := applyExtrinsicResultValidityErr.Bytes()

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bUxt)
	mockRuntimeDecoder.On("DecodeUncheckedExtrinsic", bufferUxt).Return(uxt, nil)
	mockExecutive.On("ApplyExtrinsic", uxt).Return(validityError)
	mockMemoryUtils.On("BytesToOffsetAndSize", bExtrinsicResult).Return(ptrAndSize)

	result := target.ApplyExtrinsic(dataPtr, dataLen)

	assert.Equal(t, ptrAndSize, result)
	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockExecutive.AssertCalled(t, "ApplyExtrinsic", uxt)
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", bExtrinsicResult)
}

func Test_Module_ApplyExtrinsic_DispatchError(t *testing.T) {
	target := setup()

	bufferUxt := bytes.NewBuffer(bUxt)
	dispatchErr := primitives.NewDispatchErrorBadOrigin()
	applyExtrinsicResultDispatchErr, err := primitives.NewApplyExtrinsicResult(primitives.DispatchOutcome(sc.NewVaryingData(dispatchErr)))
	assert.Nil(t, err)
	bExtrinsicResult := applyExtrinsicResultDispatchErr.Bytes()

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bUxt)
	mockRuntimeDecoder.On("DecodeUncheckedExtrinsic", bufferUxt).Return(uxt, nil)
	mockExecutive.On("ApplyExtrinsic", uxt).Return(dispatchErr)
	mockMemoryUtils.On("BytesToOffsetAndSize", bExtrinsicResult).Return(ptrAndSize)

	result := target.ApplyExtrinsic(dataPtr, dataLen)

	assert.Equal(t, ptrAndSize, result)
	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockExecutive.AssertCalled(t, "ApplyExtrinsic", uxt)
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", bExtrinsicResult)
}

func Test_Module_ApplyExtrinsic_Panics(t *testing.T) {
	target := setup()

	bufferUxt := bytes.NewBuffer(bUxt)
	expectedErr := errors.New("panic")

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bUxt)
	mockRuntimeDecoder.On("DecodeUncheckedExtrinsic", bufferUxt).Return(uxt, nil)
	mockExecutive.On("ApplyExtrinsic", uxt).Return(expectedErr)

	assert.PanicsWithValue(t,
		errPanic.Error(),
		func() { target.ApplyExtrinsic(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockExecutive.AssertCalled(t, "ApplyExtrinsic", uxt)
}

func Test_Module_ApplyExtrinsic_DecodeUncheckedExtrinsic_Panics(t *testing.T) {
	target := setup()

	bufferUxt := bytes.NewBuffer(bUxt)
	outcome, err := primitives.NewDispatchOutcome(sc.Empty{})
	assert.Nil(t, err)

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bUxt)
	mockRuntimeDecoder.On("DecodeUncheckedExtrinsic", bufferUxt).Return(uxt, errPanic)
	mockExecutive.On("ApplyExtrinsic", uxt).Return(outcome, nil)

	assert.PanicsWithValue(t,
		errPanic.Error(),
		func() { target.ApplyExtrinsic(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockExecutive.AssertNotCalled(t, "ApplyExtrinsic", uxt)
}

func Test_Module_FinalizeBlock(t *testing.T) {
	target := setup()

	parentHash := common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	stateRoot := common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355b").ToBytes()
	extrinsicsRoot := common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355a").ToBytes()
	header := primitives.Header{
		ParentHash: primitives.Blake2bHash{
			FixedSequence: sc.BytesToFixedSequenceU8(parentHash)},
		Number:         5,
		StateRoot:      primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(stateRoot)},
		ExtrinsicsRoot: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(extrinsicsRoot)},
	}

	mockExecutive.On("FinalizeBlock").Return(header, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", header.Bytes()).Return(ptrAndSize)

	result := target.FinalizeBlock()

	assert.Equal(t, ptrAndSize, result)
	mockExecutive.AssertCalled(t, "FinalizeBlock")
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", header.Bytes())
}

func Test_Module_FinalizeBlock_Panics(t *testing.T) {
	target := setup()

	mockExecutive.On("FinalizeBlock").Return(primitives.Header{}, errPanic)

	assert.PanicsWithValue(t,
		errPanic.Error(),
		func() { target.FinalizeBlock() },
	)

	mockExecutive.AssertCalled(t, "FinalizeBlock")
	mockMemoryUtils.AssertNotCalled(t, "BytesToOffsetAndSize", mock.Anything)
}

func Test_Module_InherentExtrinsics_Success(t *testing.T) {
	target := setup()

	bInherentData := sc.ToCompact(0).Bytes()
	inherentData := primitives.NewInherentData()
	bCreate := []byte{1}

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bInherentData)
	mockRuntimeExtrinsic.On("CreateInherents", *inherentData).Return(bCreate, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", bCreate).Return(ptrAndSize)

	result := target.InherentExtrinsics(dataPtr, dataLen)

	assert.Equal(t, ptrAndSize, result)
	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeExtrinsic.AssertCalled(t, "CreateInherents", *inherentData)
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", bCreate)
}

func Test_Module_InherentExtrinsics_InvalidInherentData(t *testing.T) {
	target := setup()

	bInvalidInherentData := sc.ToCompact(1).Bytes()

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bInvalidInherentData)

	assert.PanicsWithValue(t,
		io.EOF.Error(),
		func() { target.InherentExtrinsics(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeExtrinsic.AssertNotCalled(t, "CreateInherents", mock.Anything)
	mockMemoryUtils.AssertNotCalled(t, "BytesToOffsetAndSize", mock.Anything)
}

func Test_Module_InherentExtrinsics_CreateInherents_Panics(t *testing.T) {
	target := setup()

	bInherentData := sc.ToCompact(0).Bytes()
	inherentData := primitives.NewInherentData()
	bCreate := []byte{1}

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bInherentData)
	mockRuntimeExtrinsic.On("CreateInherents", *inherentData).Return(bCreate, errPanic)

	assert.PanicsWithValue(t,
		errPanic.Error(),
		func() { target.InherentExtrinsics(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeExtrinsic.AssertCalled(t, "CreateInherents", *inherentData)
	mockMemoryUtils.AssertNotCalled(t, "BytesToOffsetAndSize", mock.Anything)
}

func Test_Module_CheckInherents_Success(t *testing.T) {
	target := setup()

	block := types.NewBlock(primitives.Header{Number: 1}, sc.Sequence[primitives.UncheckedExtrinsic]{})
	bInherentData := sc.ToCompact(0).Bytes()
	bufferData := bytes.NewBuffer(bInherentData)
	inherentData := primitives.NewInherentData()
	checkResult := primitives.NewCheckInherentsResult()

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bInherentData)
	mockRuntimeDecoder.On("DecodeBlock", bufferData).Return(block, nil)
	mockRuntimeExtrinsic.On("CheckInherents", *inherentData, block).Return(checkResult, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", checkResult.Bytes()).Return(ptrAndSize)

	result := target.CheckInherents(dataPtr, dataLen)

	assert.Equal(t, ptrAndSize, result)
	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockRuntimeExtrinsic.AssertCalled(t, "CheckInherents", *inherentData, block)
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", checkResult.Bytes())
}

func Test_Module_CheckInherents_InvalidInherentData(t *testing.T) {
	target := setup()

	block := types.NewBlock(primitives.Header{Number: 1}, sc.Sequence[primitives.UncheckedExtrinsic]{})
	bytesInvalidInherentData := sc.ToCompact(1).Bytes()
	bufferData := bytes.NewBuffer(bytesInvalidInherentData)

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bytesInvalidInherentData)
	mockRuntimeDecoder.On("DecodeBlock", bufferData).Return(block, nil)

	assert.PanicsWithValue(t,
		io.EOF.Error(),
		func() { target.CheckInherents(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockRuntimeExtrinsic.AssertNotCalled(t, "CheckInherents", mock.Anything)
	mockMemoryUtils.AssertNotCalled(t, "BytesToOffsetAndSize", mock.Anything)
}

func Test_Module_CheckInherents_DecodeBlock_Panics(t *testing.T) {
	target := setup()

	block := types.NewBlock(primitives.Header{Number: 1}, sc.Sequence[primitives.UncheckedExtrinsic]{})
	bInherentData := sc.ToCompact(0).Bytes()
	bufferData := bytes.NewBuffer(bInherentData)

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bInherentData)
	mockRuntimeDecoder.On("DecodeBlock", bufferData).Return(block, errPanic)

	assert.PanicsWithValue(t,
		errPanic.Error(),
		func() { target.CheckInherents(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockRuntimeExtrinsic.AssertNotCalled(t, "CheckInherents", mock.Anything, mock.Anything)
}

func Test_Module_CheckInherents_CheckInherents_Panics(t *testing.T) {
	target := setup()

	block := types.NewBlock(primitives.Header{Number: 1}, sc.Sequence[primitives.UncheckedExtrinsic]{})
	bInherentData := sc.ToCompact(0).Bytes()
	bufferData := bytes.NewBuffer(bInherentData)
	inherentData := primitives.NewInherentData()
	checkResult := primitives.NewCheckInherentsResult()

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(bInherentData)
	mockRuntimeDecoder.On("DecodeBlock", bufferData).Return(block, nil)
	mockRuntimeExtrinsic.On("CheckInherents", *inherentData, block).Return(checkResult, errPanic)

	assert.PanicsWithValue(t,
		errPanic.Error(),
		func() { target.CheckInherents(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertExpectations(t)
	mockRuntimeExtrinsic.AssertCalled(t, "CheckInherents", *inherentData, block)
	mockMemoryUtils.AssertNotCalled(t, "BytesToOffsetAndSize", mock.Anything)
}

func Test_Module_Metadata(t *testing.T) {
	target := setup()

	blockId, _ := target.mdGenerator.GetId("block")

	expect := primitives.RuntimeApiMetadata{
		Name: ApiModuleName,
		Methods: sc.Sequence[primitives.RuntimeApiMethodMetadata]{
			primitives.RuntimeApiMethodMetadata{
				Name: "apply_extrinsic",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
					primitives.RuntimeApiMethodParamMetadata{
						Name: "Extrinsic",
						Type: sc.ToCompact(metadata.UncheckedExtrinsic),
					},
				},
				Output: sc.ToCompact(metadata.TypesApplyExtrinsicResult),
				Docs: sc.Sequence[sc.Str]{" Apply the given extrinsic.",
					"",
					" Returns an inclusion outcome which specifies if this extrinsic is included in",
					" this block or not."},
			},
			primitives.RuntimeApiMethodMetadata{
				Name:   "finalize_block",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
				Output: sc.ToCompact(metadata.Header),
				Docs:   sc.Sequence[sc.Str]{" Finish the current block."},
			},
			primitives.RuntimeApiMethodMetadata{
				Name: "inherent_extrinsics",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
					primitives.RuntimeApiMethodParamMetadata{
						Name: "inherent",
						Type: sc.ToCompact(metadata.TypesInherentData),
					},
				},
				Output: sc.ToCompact(metadata.TypesSequenceUncheckedExtrinsics),
				Docs:   sc.Sequence[sc.Str]{" Generate inherent extrinsics. The inherent data will vary from chain to chain."},
			},
			primitives.RuntimeApiMethodMetadata{
				Name: "check_inherents",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
					primitives.RuntimeApiMethodParamMetadata{
						Name: "block",
						Type: sc.ToCompact(blockId),
					},
					primitives.RuntimeApiMethodParamMetadata{
						Name: "data",
						Type: sc.ToCompact(metadata.TypesInherentData),
					},
				},
				Output: sc.ToCompact(metadata.CheckInherentsResult),
				Docs:   sc.Sequence[sc.Str]{" Check that the inherents are valid. The inherent data will vary from chain to chain."},
			},
		},
		Docs: sc.Sequence[sc.Str]{" The `BlockBuilder` api trait that provides the required functionality for building a block."},
	}

	assert.Equal(t, expect, target.Metadata())
}

func setup() Module {
	mockExecutive = new(mocks.Executive)
	mockRuntimeDecoder = new(mocks.RuntimeDecoder)
	mockRuntimeExtrinsic = new(mocks.RuntimeExtrinsic)
	mockMemoryUtils = new(mocks.MemoryTranslator)

	target := New(mockRuntimeExtrinsic, mockExecutive, mockRuntimeDecoder, mdGenerator, log.NewLogger())
	target.memUtils = mockMemoryUtils

	return target
}
