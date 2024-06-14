package parachain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/parachain_system"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	dataPtr    = int32(0)
	dataLen    = int32(1)
	ptrAndSize = int64(6)

	parachainIndex = sc.U8(20)

	hexStorageProof, _ = os.ReadFile("storage_proof_test.txt")
	stateRoot          = common.MustHexToHash("0x2b77a23bac83ac6fd7100292b7661edca65106db9d658411f24fe60c7254eb5c").ToBytes()
	parentHash         = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	extrinsicsRoot     = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355a").ToBytes()
	header             = primitives.Header{
		ParentHash: primitives.Blake2bHash{
			FixedSequence: sc.BytesToFixedSequenceU8(parentHash)},
		Number:         5,
		StateRoot:      primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(stateRoot)},
		ExtrinsicsRoot: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(extrinsicsRoot)},
		Digest:         primitives.NewDigest(sc.Sequence[primitives.DigestItem]{}),
	}
	validationParams = parachain.ValidationParams{
		ParentHead:             sc.BytesToSequenceU8(header.Bytes()),
		BlockData:              sc.Sequence[sc.U8]{},
		RelayParentBlockNumber: 1,
		RelayParentStorageRoot: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(stateRoot)},
	}
	parachainInherentData = parachain.InherentData{
		ValidationData: parachain.PersistedValidationData{
			ParentHead:             sc.BytesToSequenceU8(header.Bytes()),
			RelayParentNumber:      validationParams.RelayParentBlockNumber,
			RelayParentStorageRoot: validationParams.RelayParentStorageRoot,
			MaxPovSize:             0,
		},
		RelayChainState:    parachain.StorageProof{},
		DownwardMessages:   nil,
		HorizontalMessages: parachain.HorizontalMessages{},
	}
	collationInfo = parachain.CollationInfo{
		UpwardMessages:            sc.Sequence[parachain.UpwardMessage]{},
		HorizontalMessages:        sc.Sequence[parachain.OutboundHrmpMessage]{},
		ValidationCode:            sc.Option[sc.Sequence[sc.U8]]{},
		ProcessedDownwardMessages: 1,
		HrmpWatermark:             2,
		HeadData:                  sc.Sequence[sc.U8]{},
	}
	validationResult = parachain.ValidationResult{
		HeadData:                  collationInfo.HeadData,
		NewValidationCode:         collationInfo.ValidationCode,
		UpwardMessages:            collationInfo.UpwardMessages,
		HorizontalMessages:        collationInfo.HorizontalMessages,
		ProcessedDownwardMessages: collationInfo.ProcessedDownwardMessages,
		HrmpWatermark:             collationInfo.HrmpWatermark,
	}
	errPanic = errors.New("panic")
)

var (
	mockHashing            *mocks.IoHashing
	mockCall               *mocks.Call
	mockUncheckedExtrinsic *mocks.UncheckedExtrinsic
	mockBlock              *mocks.Block
	mockParachainSystem    *mocks.ParachainSystemModule
	mockBlockExecutor      *mocks.BlockExecutor
	mockRuntimeDecoder     *mocks.RuntimeDecoder
	mockHostEnvironment    *mocks.HostEnvironment
	mockMemoryUtils        *mocks.MemoryTranslator
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

func Test_Module_ValidateBlock(t *testing.T) {
	target := setup()

	bytesStorageProof, err := hex.DecodeString(string(hexStorageProof))
	assert.NoError(t, err)
	storageProof, err := parachain.DecodeStorageProof(bytes.NewBuffer(bytesStorageProof))
	assert.NoError(t, err)

	blockData := parachain.BlockData{
		Block:        mockBlock,
		CompactProof: storageProof,
	}

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(validationParams.Bytes(), nil)
	mockRuntimeDecoder.On("DecodeParachainBlockData", validationParams.BlockData).Return(blockData, nil)
	mockHashing.On("Blake256", header.Bytes()).Return(header.ParentHash.Bytes())
	mockBlock.On("Header").Return(header)
	mockBlock.On("Extrinsics").Return(sc.Sequence[primitives.UncheckedExtrinsic]{mockUncheckedExtrinsic})
	mockUncheckedExtrinsic.On("IsSigned").Return(false)
	mockUncheckedExtrinsic.On("Function").Return(mockCall)
	mockParachainSystem.On("GetIndex").Return(parachainIndex)
	mockCall.On("ModuleIndex").Return(parachainIndex)
	mockCall.On("FunctionIndex").Return(sc.U8(parachain_system.FunctionSetValidationData))
	mockCall.On("Args").Return(sc.NewVaryingData(parachainInherentData))
	mockHostEnvironment.On("SetTrieState", mock.Anything)
	mockBlockExecutor.On("ExecuteBlock", blockData.Block).Return(nil)
	mockParachainSystem.On("CollectCollationInfo", header).Return(collationInfo, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", validationResult.Bytes()).Return(ptrAndSize)

	result := target.ValidateBlock(dataPtr, dataLen)
	assert.Equal(t, ptrAndSize, result)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockRuntimeDecoder.AssertCalled(t, "DecodeParachainBlockData", validationParams.BlockData)
	mockHashing.AssertCalled(t, "Blake256", header.Bytes())
	mockBlock.AssertCalled(t, "Header")
	mockBlock.AssertCalled(t, "Extrinsics")
	mockUncheckedExtrinsic.AssertCalled(t, "IsSigned")
	mockUncheckedExtrinsic.AssertCalled(t, "Function")
	mockParachainSystem.AssertCalled(t, "GetIndex")
	mockCall.AssertCalled(t, "ModuleIndex")
	mockCall.AssertCalled(t, "FunctionIndex")
	mockCall.AssertCalled(t, "Args")
	mockHostEnvironment.AssertCalled(t, "SetTrieState", mock.Anything)
	mockBlockExecutor.AssertCalled(t, "ExecuteBlock", blockData.Block)
	mockParachainSystem.AssertCalled(t, "CollectCollationInfo", header)
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", validationResult.Bytes())
}

func setup() Module {
	mockHashing = new(mocks.IoHashing)
	mockCall = new(mocks.Call)
	mockUncheckedExtrinsic = new(mocks.UncheckedExtrinsic)
	mockBlock = new(mocks.Block)
	mockParachainSystem = new(mocks.ParachainSystemModule)
	mockBlockExecutor = new(mocks.BlockExecutor)
	mockRuntimeDecoder = new(mocks.RuntimeDecoder)
	mockHostEnvironment = new(mocks.HostEnvironment)
	mockMemoryUtils = new(mocks.MemoryTranslator)

	target := New(mockParachainSystem, mockBlockExecutor, mockRuntimeDecoder, mockHostEnvironment, log.NewLogger())
	target.hashing = mockHashing
	target.memUtils = mockMemoryUtils

	return target
}
