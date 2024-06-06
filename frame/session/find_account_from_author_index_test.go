package session

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/mocks"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	validatorAccountIds = sc.Sequence[primitives.AccountId]{constants.ZeroAccountId, constants.OneAccountId, constants.TwoAccountId}
	authorityIndex      = sc.U32(1)
	output              = sc.NewFixedSequence(32, make([]sc.U8, 32)...)
	proof               = sc.NewFixedSequence(64, make([]sc.U8, 64)...)
	vrfSignature        = primitives.VrfSignature{PreOutput: output, Proof: proof}

	primaryPreDigest = babetypes.PrimaryPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           sc.U64(2),
		VrfSignature:   vrfSignature,
	}

	preDigest = babetypes.PreDigest{VaryingData: sc.NewVaryingData(babetypes.Primary, primaryPreDigest)}

	engineId = [4]byte{'B', 'A', 'B', 'E'}

	digestsPreRuntime = sc.Sequence[primitives.DigestPreRuntime]{
		{
			ConsensusEngineId: sc.BytesToFixedSequenceU8(engineId[:]),
			Message:           sc.BytesToSequenceU8(preDigest.Bytes()),
		},
	}
)

var (
	mockSessionModule *mocks.SessionModule
	mockBabeModule    *mocks.BabeModule
)

var target FindAccountFromAuthorIndex

func setup() {
	mockSessionModule = new(mocks.SessionModule)
	mockBabeModule = new(mocks.BabeModule)

	target = NewFindAccountFromAuthorIndex(mockSessionModule, mockBabeModule)
}

func Test_NewFindAccountFromAuthorIndex(t *testing.T) {
	setup()

	assert.Equal(t, FindAccountFromAuthorIndex{sessionModule: mockSessionModule, authorFinder: mockBabeModule}, target)
}

func Test_FindAuthor(t *testing.T) {
	setup()

	mockBabeModule.On("FindAuthor", digestsPreRuntime).Return(sc.NewOption[sc.U32](authorityIndex), nil)
	mockSessionModule.On("Validators").Return(validatorAccountIds, nil)

	result, err := target.FindAuthor(digestsPreRuntime)

	assert.NoError(t, err)
	assert.Equal(t, sc.NewOption[primitives.AccountId](constants.OneAccountId), result)

	mockBabeModule.AssertCalled(t, "FindAuthor", digestsPreRuntime)
	mockSessionModule.AssertCalled(t, "Validators")
}

func Test_FindAuthor_AuthorIndex_Not_Found(t *testing.T) {
	setup()

	digests := sc.Sequence[primitives.DigestPreRuntime]{
		{
			ConsensusEngineId: sc.BytesToFixedSequenceU8(engineId[:]),
			Message:           sc.BytesToSequenceU8([]byte{}),
		},
	}

	mockBabeModule.On("FindAuthor", digests).Return(sc.NewOption[sc.U32](nil), nil)

	result, err := target.FindAuthor(digests)

	assert.Equal(t, errAuthorNotFound, err)
	assert.Equal(t, sc.NewOption[primitives.AccountId](nil), result)

	mockBabeModule.AssertCalled(t, "FindAuthor", digests)
	mockSessionModule.AssertNotCalled(t, "Validators")
}

func Test_FindAuthor_Fails_To_Get_AuthorIndex(t *testing.T) {
	setup()

	someError := errors.New("author index error")

	mockBabeModule.On("FindAuthor", digestsPreRuntime).Return(sc.NewOption[sc.U32](nil), someError)

	result, err := target.FindAuthor(digestsPreRuntime)

	assert.Equal(t, someError, err)
	assert.Equal(t, sc.NewOption[primitives.AccountId](nil), result)

	mockBabeModule.AssertCalled(t, "FindAuthor", digestsPreRuntime)
	mockSessionModule.AssertNotCalled(t, "Validators")
}

func Test_FindAuthor_Fails_To_Get_Validators(t *testing.T) {
	setup()

	someError := errors.New("get validators error")

	mockBabeModule.On("FindAuthor", digestsPreRuntime).Return(sc.NewOption[sc.U32](authorityIndex), nil)
	mockSessionModule.On("Validators").Return(sc.Sequence[primitives.AccountId]{}, someError)

	result, err := target.FindAuthor(digestsPreRuntime)

	assert.Equal(t, someError, err)
	assert.Equal(t, sc.NewOption[primitives.AccountId](nil), result)

	mockBabeModule.AssertCalled(t, "FindAuthor", digestsPreRuntime)
	mockSessionModule.AssertCalled(t, "Validators")
}

func Test_FindAuthor_Author_Index_Out_Of_Bounds(t *testing.T) {
	setup()

	mockBabeModule.On("FindAuthor", digestsPreRuntime).Return(sc.NewOption[sc.U32](sc.U32(3)), nil)
	mockSessionModule.On("Validators").Return(validatorAccountIds, nil)

	result, err := target.FindAuthor(digestsPreRuntime)

	assert.Equal(t, errAuthorIndexOutOfBounds, err)
	assert.Equal(t, sc.NewOption[primitives.AccountId](nil), result)

	mockBabeModule.AssertCalled(t, "FindAuthor", digestsPreRuntime)
	mockSessionModule.AssertCalled(t, "Validators")
}
