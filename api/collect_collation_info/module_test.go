package collect_collation_info

import (
	"errors"
	"io"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	dataPtr    = int32(0)
	dataLen    = int32(1)
	ptrAndSize = int64(6)

	parentHash     = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	stateRoot      = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355b").ToBytes()
	extrinsicsRoot = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355a").ToBytes()
	header         = primitives.Header{
		ParentHash: primitives.Blake2bHash{
			FixedSequence: sc.BytesToFixedSequenceU8(parentHash)},
		Number:         5,
		StateRoot:      primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(stateRoot)},
		ExtrinsicsRoot: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(extrinsicsRoot)},
		Digest:         primitives.NewDigest(sc.Sequence[primitives.DigestItem]{}),
	}
	collationInfo = parachain.CollationInfo{
		UpwardMessages:            sc.Sequence[parachain.UpwardMessage]{},
		HorizontalMessages:        sc.Sequence[parachain.OutboundHrmpMessage]{},
		ValidationCode:            sc.Option[sc.Sequence[sc.U8]]{},
		ProcessedDownwardMessages: 0,
		HrmpWatermark:             0,
		HeadData:                  sc.Sequence[sc.U8]{},
	}
	errPanic = errors.New("panic")
)

var (
	mockParachainSystemModule *mocks.ParachainSystemModule
	mockMemoryUtils           *mocks.MemoryTranslator
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

func Test_Module_CollectCollationInfo(t *testing.T) {
	target := setup()

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(header.Bytes())
	mockParachainSystemModule.On("CollectCollationInfo", header).Return(collationInfo, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", collationInfo.Bytes()).Return(ptrAndSize)

	result := target.CollectCollationInfo(dataPtr, dataLen)

	assert.Equal(t, ptrAndSize, result)
	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockParachainSystemModule.AssertCalled(t, "CollectCollationInfo", header)
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", collationInfo.Bytes())
}

func Test_Module_CollectCollationInfo_Header_Panics(t *testing.T) {
	target := setup()

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return([]byte{})

	assert.PanicsWithValue(t,
		io.EOF.Error(),
		func() { target.CollectCollationInfo(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
}

func Test_Module_CollectCollationInfo_Panics(t *testing.T) {
	target := setup()

	expectedErr := errors.New("panic")

	mockMemoryUtils.On("GetWasmMemorySlice", dataPtr, dataLen).Return(header.Bytes())
	mockParachainSystemModule.On("CollectCollationInfo", header).Return(collationInfo, expectedErr)

	assert.PanicsWithValue(t,
		errPanic.Error(),
		func() { target.CollectCollationInfo(dataPtr, dataLen) },
	)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", dataPtr, dataLen)
	mockParachainSystemModule.On("CollectCollationInfo", header).Return(collationInfo, expectedErr)
}

func Test_Module_Metadata(t *testing.T) {
	target := setup()

	expect := primitives.RuntimeApiMetadata{
		Name: ApiModuleName,
		Methods: sc.Sequence[primitives.RuntimeApiMethodMetadata]{
			primitives.RuntimeApiMethodMetadata{
				Name: "collect_collation_info",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
					primitives.RuntimeApiMethodParamMetadata{
						Name: "header",
						Type: sc.ToCompact(metadata.Header),
					},
				},
				Output: sc.ToCompact(metadata.TypesParachainValidationResult),
				Docs:   sc.Sequence[sc.Str]{"Returns the collation information by the block header."},
			},
		},
		Docs: sc.Sequence[sc.Str]{" The collect collation info api."},
	}

	assert.Equal(t, expect, target.Metadata())
}

func setup() Module {
	mockParachainSystemModule = new(mocks.ParachainSystemModule)
	mockMemoryUtils = new(mocks.MemoryTranslator)

	target := New(mockParachainSystemModule, log.NewLogger())
	target.memUtils = mockMemoryUtils

	return target
}
