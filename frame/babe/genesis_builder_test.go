package babe

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
)

var (
	moduleId                      = sc.U8(2)
	maxAuthorities         sc.U32 = 100
	timestampMinimumPeriod sc.U64 = 1 * 1_000
	epochDuration                 = constants.EpochDurationInSlots
	genesisEpochConfig            = BabeEpochConfiguration{C: constants.PrimaryProbability, AllowedSlots: NewPrimaryAndSecondaryVRFSlots()}
	pubKey, _                     = types.NewSr25519PublicKey(sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)...)
	authorities                   = sc.Sequence[Authority]{Authority{Key: pubKey}}
)

var (
	expectedGenesisConfig = GenesisConfig{
		Authorities: authorities,
		EpochConfig: genesisEpochConfig,
	}

	expectedJson = "{\"babe\":{\"authorities\":[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"],\"epochConfig\":{\"c\":[1,4],\"allowed_slots\":\"PrimaryAndSecondaryVRFSlots\"}}}"
)

var (
	mockStorageSegmentIndex    *mocks.StorageValue[sc.U32]
	mockStorageAuthorities     *mocks.StorageValue[sc.Sequence[Authority]]
	mockStorageNextAuthorities *mocks.StorageValue[sc.Sequence[Authority]]
	mockStorageEpochConfig     *mocks.StorageValue[BabeEpochConfiguration]
)

var target module

func Test_Babe_CreateDefaultConfig(t *testing.T) {
	target := setupModule()

	gcJson, err := target.CreateDefaultConfig()

	assert.NoError(t, err)
	assert.Equal(t, expectedJson, string(gcJson))
}

func Test_Babe_BuildConfig(t *testing.T) {
	target := setupModule()

	mockStorageSegmentIndex.On("Put", sc.U32(0)).Return()
	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](nil), nil)
	mockStorageAuthorities.On("Put", authorities).Return()
	mockStorageNextAuthorities.On("Put", authorities).Return()
	mockStorageEpochConfig.On("Put", genesisEpochConfig).Return()

	err := target.BuildConfig([]byte(expectedJson))

	assert.NoError(t, err)
	mockStorageSegmentIndex.AssertCalled(t, "Put", sc.U32(0))
	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertCalled(t, "Put", authorities)
	mockStorageEpochConfig.AssertCalled(t, "Put", genesisEpochConfig)
}

// TODO: add more test cases
// duplicate genesis address
// invalid ss58 address
// zero authorities
// storage authorities DecodeLen error
// storage authorities DecodeLen has value
// authorities exceed max authorities

func setupModule() module {
	mockStorageSegmentIndex = new(mocks.StorageValue[sc.U32])
	mockStorageAuthorities = new(mocks.StorageValue[sc.Sequence[Authority]])
	mockStorageNextAuthorities = new(mocks.StorageValue[sc.Sequence[Authority]])
	mockStorageEpochConfig = new(mocks.StorageValue[BabeEpochConfiguration])

	config := NewConfig(
		types.PublicKeySr25519,
		genesisEpochConfig,
		epochDuration,
		timestampMinimumPeriod,
		maxAuthorities,
	)

	target := New(moduleId, config).(module)
	target.storage.SegmentIndex = mockStorageSegmentIndex
	target.storage.Authorities = mockStorageAuthorities
	target.storage.NextAuthorities = mockStorageNextAuthorities
	target.storage.EpochConfig = mockStorageEpochConfig

	return target
}
