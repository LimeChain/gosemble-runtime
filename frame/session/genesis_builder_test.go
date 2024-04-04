package session

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	bytesAccount    = []byte{0xd4, 0x35, 0x93, 0xc7, 0x15, 0xfd, 0xd3, 0x1c, 0x61, 0x14, 0x1a, 0xbd, 0x4, 0xa9, 0x9f, 0xd6, 0x82, 0x2c, 0x85, 0x58, 0x85, 0x4c, 0xcd, 0xe3, 0x9a, 0x56, 0x84, 0xe7, 0xa5, 0x6d, 0xa2, 0x7d}
	fixedSeqAccount = sc.BytesToFixedSequenceU8(bytesAccount)
	accountId, _    = primitives.NewAccountId(fixedSeqAccount...)
)

func Test_GenesisConfig_CreateDefaultConfig(t *testing.T) {
	target := setupModule()

	expectedGc := []byte("{\"session\":{\"keys\":[]}}")

	gc, err := target.CreateDefaultConfig()
	assert.NoError(t, err)
	assert.Equal(t, expectedGc, gc)
}

func Test_GenesisConfig_BuildConfig(t *testing.T) {
	key := sc.BytesToSequenceU8(bytesAccount)
	sessionKey := primitives.SessionKey{
		Key:    key,
		TypeId: sc.FixedSequence[sc.U8]{'t', 'e', 's', 't'},
	}
	nextKeys := sc.FixedSequence[primitives.Sr25519PublicKey]{
		primitives.Sr25519PublicKey{FixedSequence: fixedSeqAccount},
	}
	queuedKeys := sc.Sequence[queuedKey]{
		queuedKey{
			Validator: accountId,
			Keys:      sc.Sequence[primitives.SessionKey](nil),
		},
	}
	validators := sc.Sequence[primitives.AccountId]{
		accountId,
	}
	gcJson := "{\"session\":{\"keys\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\", \"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",{\"test\":\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"}]]}}"
	target := setupModule()

	mockStorageNextKeys.On("Get", accountId).Return(sc.FixedSequence[primitives.Sr25519PublicKey]{}, nil)
	mockSessionHandler.On("KeyTypeIds").Return(keyTypeIds)
	mockStorageKeyOwner.On("Get", sessionKey).Return(constants.ZeroAccountId, nil)
	mockSystemModule.On("IncConsumers", accountId).Return(nil)
	mockStorageKeyOwner.On("Put", sessionKey, accountId).Return()
	mockStorageNextKeys.On("Put", accountId, nextKeys).Return()

	mockSystemModule.On("IncConsumersWithoutLimit", accountId).Return(nil)
	mockSessionHandler.On("OnGenesisSession", queuedKeys).Return(nil)
	mockStorageValidators.On("Put", validators).Return()
	mockStorageQueuedKeys.On("Put", queuedKeys).Return()

	err := target.BuildConfig([]byte(gcJson))
	assert.Nil(t, err)

	mockStorageNextKeys.AssertCalled(t, "Get", accountId)
	mockSessionHandler.AssertCalled(t, "KeyTypeIds")
	mockStorageKeyOwner.AssertCalled(t, "Get", sessionKey)
	mockStorageKeyOwner.AssertCalled(t, "Put", sessionKey, accountId)
	mockStorageNextKeys.AssertCalled(t, "Put", accountId, nextKeys)

	mockSystemModule.AssertCalled(t, "IncConsumersWithoutLimit", accountId)
	mockSessionHandler.AssertCalled(t, "OnGenesisSession", queuedKeys)
	mockStorageValidators.AssertCalled(t, "Put", validators)
	mockStorageQueuedKeys.AssertCalled(t, "Put", queuedKeys)
}
