package aura_ext

import (
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/aura"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	stateRoot      = common.MustHexToHash("0x2b77a23bac83ac6fd7100292b7661edca65106db9d658411f24fe60c7254eb5c").ToBytes()
	parentHash     = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	extrinsicsRoot = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355a").ToBytes()
	sealMessage    = sc.Sequence[sc.U8]{'t', 'e', 's', 't'}
	seal           = primitives.NewDigestItemSeal(sc.BytesToFixedSequenceU8(aura.EngineId[:]), sealMessage)
	header         = primitives.Header{
		ParentHash: primitives.Blake2bHash{
			FixedSequence: sc.BytesToFixedSequenceU8(parentHash)},
		Number:         5,
		StateRoot:      primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(stateRoot)},
		ExtrinsicsRoot: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(extrinsicsRoot)},
		Digest: primitives.NewDigest(sc.Sequence[primitives.DigestItem]{
			seal,
		}),
	}
	authority   = primitives.Sr25519PublicKey{FixedSequence: constants.OneAddress.FixedSequence}
	authorities = sc.Sequence[primitives.Sr25519PublicKey]{
		authority,
	}
	authorIndex       = sc.U32(0)
	optionAuthorIndex = sc.NewOption[sc.U32](authorIndex)
	preHash           = []byte{'p', 'r', 'e', 'h', 'a', 's', 'h'}
)

var (
	mockCrypto          *mocks.IoCrypto
	mockIoHashing       *mocks.IoHashing
	mockExecutiveModule *mocks.Executive
	mockBlock           *mocks.Block
)

func Test_BlockExecutor_ExecuteBlock(t *testing.T) {
	target := setupBlockExecutor()
	hashedHeader := header
	hashedHeader.Digest = primitives.NewDigest(sc.Sequence[primitives.DigestItem]{})
	expectBlock := types.NewBlock(hashedHeader, sc.Sequence[primitives.UncheckedExtrinsic]{})

	mockBlock.On("Header").Return(header)
	mockAuthorities.On("Get").Return(authorities, nil)
	mockAuraModule.On("FindAuthor", sc.Sequence[primitives.DigestPreRuntime]{}).Return(optionAuthorIndex, nil)
	mockIoHashing.On("Blake256", hashedHeader.Bytes()).Return(preHash)
	mockCrypto.On("Sr25519Verify", sc.SequenceU8ToBytes(sealMessage), preHash, sc.FixedSequenceU8ToBytes(authority.FixedSequence)).Return(true)
	mockBlock.On("Extrinsics").Return(sc.Sequence[primitives.UncheckedExtrinsic]{})
	mockExecutiveModule.On("ExecuteBlock", expectBlock).Return(nil)

	err := target.ExecuteBlock(mockBlock)
	assert.Nil(t, err)

	mockBlock.AssertCalled(t, "Header")
	mockAuthorities.AssertCalled(t, "Get")
	mockAuraModule.AssertCalled(t, "FindAuthor", sc.Sequence[primitives.DigestPreRuntime]{})
	mockIoHashing.AssertCalled(t, "Blake256", hashedHeader.Bytes())
	mockCrypto.AssertCalled(t, "Sr25519Verify", sc.SequenceU8ToBytes(sealMessage), preHash, sc.FixedSequenceU8ToBytes(authority.FixedSequence))
	mockBlock.AssertCalled(t, "Extrinsics")
	mockExecutiveModule.AssertCalled(t, "ExecuteBlock", expectBlock)
}

func setupBlockExecutor() blockExecutor {
	mockCrypto = new(mocks.IoCrypto)
	mockIoHashing = new(mocks.IoHashing)
	mockExecutiveModule = new(mocks.Executive)
	mockBlock = new(mocks.Block)

	module := setupModule()

	target := NewBlockExecutor(module, mockExecutiveModule).(blockExecutor)
	target.ioCrypto = mockCrypto
	target.ioHashing = mockIoHashing

	return target
}
