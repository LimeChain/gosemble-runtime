package babe

import (
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/frame/session"
	"github.com/LimeChain/gosemble/mocks"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
)

var (
	moduleId = sc.U8(2)

	now = sc.U64(123456)

	timestampMinimumPeriod sc.U64 = 1 * 1_000
	epochIndex                    = sc.U64(2)
	genesisSlot                   = sc.U64(100)
	epochDuration                 = constants.EpochDurationInSlots
	epochConfig                   = babetypes.EpochConfiguration{
		C:            constants.PrimaryProbability,
		AllowedSlots: babetypes.NewPrimarySlots(),
	}
	slot                   = sc.U64(130)
	maxAuthorities  sc.U32 = 50
	authorityIndex         = sc.U32(1)
	pubKey, _              = types.NewSr25519PublicKey(sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)...)
	aliceAuthority         = babetypes.Authority{Key: pubKey}
	authorities            = sc.Sequence[babetypes.Authority]{aliceAuthority}
	nextAuthorities        = sc.Sequence[babetypes.Authority]{aliceAuthority}

	output = sc.NewFixedSequence(32, make([]sc.U8, 32)...)
	proof  = sc.NewFixedSequence(64, make([]sc.U8, 64)...)

	vrfSignature = types.VrfSignature{
		PreOutput: output,
		Proof:     proof,
	}

	randomness     = sc.NewFixedSequence[sc.U8](32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	nextRandomness = sc.NewFixedSequence[sc.U8](32, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1)

	preDigest = PreDigest{sc.NewVaryingData(Primary, primaryPreDigest)}

	digestsPreRuntime = sc.Sequence[types.DigestPreRuntime]{
		{
			ConsensusEngineId: sc.BytesToFixedSequenceU8(EngineId[:]),
			Message:           sc.BytesToSequenceU8(preDigest.Bytes()),
		},
	}

	digestsPreRuntimeInvalidMessage = sc.Sequence[types.DigestPreRuntime]{
		{
			ConsensusEngineId: sc.BytesToFixedSequenceU8(EngineId[:]),
			Message:           sc.BytesToSequenceU8(primaryPreDigest.Bytes()),
		},
	}

	digestsPreRuntimeTestEngineId = sc.Sequence[types.DigestPreRuntime]{
		{
			ConsensusEngineId: sc.BytesToFixedSequenceU8([]byte{'T', 'E', 'S', 'T'}),
			Message:           sc.BytesToSequenceU8(preDigest.Bytes()),
		},
	}

	nextEpochDataLog = types.NewDigestItemConsensusMessage(
		sc.BytesToFixedSequenceU8(EngineId[:]),
		sc.BytesToSequenceU8(NewNextEpochDataConsensusLog(
			NextEpochDescriptor{Authorities: authorities, Randomness: randomness},
		).Bytes()),
	)

	nextConfigDescriptor = NextConfigDescriptor{
		V1: babetypes.EpochConfiguration{
			C: types.RationalValue{
				Numerator:   3,
				Denominator: 5,
			},
			AllowedSlots: babetypes.NewPrimaryAndSecondaryVRFSlots(),
		},
	}

	skippedEpoch = babetypes.SkippedEpoch{}
)

var (
	expectedGenesisConfig = GenesisConfig{
		Authorities: authorities,
		EpochConfig: epochConfig,
	}

	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
)

var (
	mockSystemDigestFn = func() (primitives.Digest, error) {
		items := sc.Sequence[types.DigestItem]{
			types.NewDigestItemPreRuntime(
				sc.BytesToFixedSequenceU8(EngineId[:]),
				sc.BytesToSequenceU8(preDigest.Bytes()),
			),
		}
		return types.NewDigest(items), nil
	}
)

var (
	mockStorageAuthorities              *mocks.StorageValue[sc.Sequence[babetypes.Authority]]
	mockStorageNextEpochConfig          *mocks.StorageValue[babetypes.EpochConfiguration]
	mockStorageCurrentSlot              *mocks.StorageValue[babetypes.Slot]
	mockStorageRandomness               *mocks.StorageValue[babetypes.Randomness]
	mockStorageSegmentIndex             *mocks.StorageValue[sc.U32]
	mockStorageEpochConfig              *mocks.StorageValue[babetypes.EpochConfiguration]
	mockStorageEpochIndex               *mocks.StorageValue[sc.U64]
	mockStorageEpochStart               *mocks.StorageValue[babetypes.EpochStartBlocks]
	mockStorageGenesisSlot              *mocks.StorageValue[babetypes.Slot]
	mockStorageNextAuthorities          *mocks.StorageValue[sc.Sequence[babetypes.Authority]]
	mockStorageNextRandomness           *mocks.StorageValue[babetypes.Randomness]
	mockStoragePendingEpochConfigChange *mocks.StorageValue[NextConfigDescriptor]
	mockStorageInitialized              *mocks.StorageValue[sc.Option[PreDigest]]
	mockStorageLateness                 *mocks.StorageValue[sc.U64]
	mockStorageSkippedEpochs            *mocks.StorageValue[sc.FixedSequence[babetypes.SkippedEpoch]]
)

var (
	mockIoHashing *mocks.IoHashing
)

var (
	mockSystemModule       *mocks.SystemModule
	mockSessionModule      *mocks.SessionModule
	mockEpochChangeTrigger *mocks.EpochChangeTrigger
)

var target module

func setupModule() module {
	mockStorageAuthorities = new(mocks.StorageValue[sc.Sequence[babetypes.Authority]])
	mockStorageNextEpochConfig = new(mocks.StorageValue[babetypes.EpochConfiguration])
	mockStorageCurrentSlot = new(mocks.StorageValue[babetypes.Slot])
	mockStorageRandomness = new(mocks.StorageValue[babetypes.Randomness])
	mockStorageSegmentIndex = new(mocks.StorageValue[sc.U32])
	mockStorageEpochConfig = new(mocks.StorageValue[babetypes.EpochConfiguration])
	mockStorageEpochIndex = new(mocks.StorageValue[sc.U64])
	mockStorageEpochStart = new(mocks.StorageValue[babetypes.EpochStartBlocks])
	mockStorageGenesisSlot = new(mocks.StorageValue[babetypes.Slot])
	mockStorageNextAuthorities = new(mocks.StorageValue[sc.Sequence[babetypes.Authority]])
	mockStorageNextRandomness = new(mocks.StorageValue[babetypes.Randomness])
	mockStoragePendingEpochConfigChange = new(mocks.StorageValue[NextConfigDescriptor])
	mockStorageInitialized = new(mocks.StorageValue[sc.Option[PreDigest]])
	mockStorageLateness = new(mocks.StorageValue[sc.U64])
	mockStorageSkippedEpochs = new(mocks.StorageValue[sc.FixedSequence[babetypes.SkippedEpoch]])

	mockSystemModule = new(mocks.SystemModule)
	mockSessionModule = new(mocks.SessionModule)
	mockEpochChangeTrigger = new(mocks.EpochChangeTrigger)

	mockIoHashing = new(mocks.IoHashing)

	config := NewConfig(
		types.PublicKeySr25519,
		epochConfig,
		epochDuration,
		mockEpochChangeTrigger,
		*new(session.Module),
		maxAuthorities,
		timestampMinimumPeriod,
		mockSystemDigestFn,
		mockSystemModule,
	)

	target = New(moduleId, config, types.NewMetadataTypeGenerator(), log.NewLogger()).(module)

	target.storage.Authorities = mockStorageAuthorities
	target.storage.NextEpochConfig = mockStorageNextEpochConfig
	target.storage.CurrentSlot = mockStorageCurrentSlot
	target.storage.Randomness = mockStorageRandomness
	target.storage.SegmentIndex = mockStorageSegmentIndex
	target.storage.EpochConfig = mockStorageEpochConfig
	target.storage.EpochIndex = mockStorageEpochIndex
	target.storage.EpochStart = mockStorageEpochStart
	target.storage.GenesisSlot = mockStorageGenesisSlot
	target.storage.NextAuthorities = mockStorageNextAuthorities
	target.storage.NextRandomness = mockStorageNextRandomness
	target.storage.PendingEpochConfigChange = mockStoragePendingEpochConfigChange
	target.storage.Initialized = mockStorageInitialized
	target.storage.Lateness = mockStorageLateness

	target.ioHashing = mockIoHashing

	return target
}

func Test_Module_GetIndex(t *testing.T) {
	target := setupModule()
	assert.Equal(t, sc.U8(moduleId), target.GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	target := setupModule()
	functions := target.Functions()

	assert.Equal(t, 1, len(functions))
}

func Test_Module_PreDispatch(t *testing.T) {
	target := setupModule()
	mockCall := new(mocks.Call)

	result, err := target.PreDispatch(mockCall)

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	target := setupModule()
	mockCall := new(mocks.Call)

	result, err := target.ValidateUnsigned(primitives.TransactionSource{}, mockCall)

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Module_StorageAuthorities(t *testing.T) {
	target := setupModule()

	mockStorageAuthorities.On("Get").Return(authorities, nil)

	result, err := target.StorageAuthorities()

	mockStorageAuthorities.AssertCalled(t, "Get")

	assert.NoError(t, err)
	assert.Equal(t, authorities, result)
}

func Test_Module_StorageRandomness(t *testing.T) {
	target := setupModule()

	mockStorageRandomness.On("Get").Return(randomness, nil)

	result, err := target.StorageRandomness()

	mockStorageRandomness.AssertCalled(t, "Get")

	assert.NoError(t, err)
	assert.Equal(t, randomness, result)
}

func Test_Module_StorageSegmentIndexSet(t *testing.T) {
	target := setupModule()

	value := sc.U32(100)

	mockStorageSegmentIndex.On("Put", value).Return(nil)

	target.StorageSegmentIndexSet(value)

	mockStorageSegmentIndex.AssertCalled(t, "Put", value)
}

func Test_Module_StorageEpochConfig(t *testing.T) {
	target := setupModule()

	mockStorageEpochConfig.On("Get").Return(epochConfig, nil)

	result, err := target.StorageEpochConfig()

	mockStorageEpochConfig.AssertCalled(t, "Get")

	assert.NoError(t, err)
	assert.Equal(t, epochConfig, result)
}

func Test_Module_StorageEpochConfigSet(t *testing.T) {
	target := setupModule()

	mockStorageEpochConfig.On("Put", epochConfig)

	target.StorageEpochConfigSet(epochConfig)

	mockStorageEpochConfig.AssertCalled(t, "Put", epochConfig)
}

func Test_Module_FindAuthor(t *testing.T) {
	target := setupModule()

	index, err := target.FindAuthor(digestsPreRuntime)
	assert.NoError(t, err)

	assert.Equal(t, sc.Bool(true), index.HasValue)
	assert.Equal(t, authorityIndex, index.Value)
}

func Test_Module_FindAuthor_Invalid_Message(t *testing.T) {
	target := setupModule()

	index, err := target.FindAuthor(digestsPreRuntimeInvalidMessage)
	assert.Error(t, io.EOF, err)

	assert.Equal(t, sc.Bool(false), index.HasValue)
}

func Test_Module_FindAuthor_Different_EngineId(t *testing.T) {
	target := setupModule()

	index, err := target.FindAuthor(digestsPreRuntimeTestEngineId)
	assert.NoError(t, err)

	assert.Equal(t, sc.Bool(false), index.HasValue)
}

func Test_Module_ShouldEndSession(t *testing.T) {
	target := setupModule()

	genesisSlot := sc.U64(0)
	currentSlot := sc.U64(126)
	lateness := sc.U64(3)

	mockStorageInitialized.On("Get").Return(sc.NewOption[PreDigest](nil), nil)
	mockStorageGenesisSlot.On("Get").Return(genesisSlot, nil)

	mockStorageGenesisSlot.On("Put", slot).Return(nil)
	mockStorageGenesisSlot.On("Get").Return(genesisSlot, nil)
	mockStorageAuthorities.On("Get").Return(authorities, nil)
	mockStorageRandomness.On("Get").Return(randomness, nil)
	mockSystemModule.On("DepositLog", nextEpochDataLog).Return(nil)

	mockStorageCurrentSlot.On("Get").Return(currentSlot, nil)
	mockStorageLateness.On("Put", lateness).Return(nil)
	mockStorageCurrentSlot.On("Put", slot).Return(nil)
	mockStorageInitialized.On("Put", sc.NewOption[PreDigest](preDigest)).Return(nil)
	mockEpochChangeTrigger.On("Trigger", now).Return(nil)

	mockStorageCurrentSlot.On("Get").Return(slot, nil)
	mockStorageEpochIndex.On("Get").Return(epochIndex, nil)
	mockStorageGenesisSlot.On("Get").Return(slot, nil)

	result := target.ShouldEndSession(now)

	mockStorageAuthorities.AssertCalled(t, "Get")
	mockStorageRandomness.AssertCalled(t, "Get")
	mockSystemModule.AssertCalled(t, "DepositLog", nextEpochDataLog)
	mockStorageLateness.AssertCalled(t, "Put", lateness)
	mockEpochChangeTrigger.AssertCalled(t, "Trigger", now)
	mockStorageEpochIndex.AssertCalled(t, "Get")

	mockStorageInitialized.AssertCalled(t, "Get")
	mockStorageInitialized.AssertCalled(t, "Put", sc.NewOption[PreDigest](preDigest))

	mockStorageCurrentSlot.AssertNumberOfCalls(t, "Get", 2)
	mockStorageCurrentSlot.AssertCalled(t, "Put", slot)

	mockStorageGenesisSlot.AssertNumberOfCalls(t, "Get", 3)
	mockStorageGenesisSlot.AssertCalled(t, "Put", slot)

	assert.False(t, result)
}

func Test_Module_ShouldEpochChange(t *testing.T) {
	target := setupModule()

	now := sc.U64(123456)

	mockStorageCurrentSlot.On("Get").Return(sc.U64(601), nil)
	mockStorageEpochIndex.On("Get").Return(sc.U64(2), nil)
	mockStorageGenesisSlot.On("Get").Return(sc.U64(1), nil)

	result := target.ShouldEpochChange(now)

	assert.True(t, result)

	mockStorageCurrentSlot.AssertCalled(t, "Get")
	mockStorageEpochIndex.AssertCalled(t, "Get")
	mockStorageGenesisSlot.AssertCalled(t, "Get")
}

func Test_CurrentEpochStart(t *testing.T) {
	target := setupModule()

	mockStorageEpochIndex.On("Get").Return(epochIndex, nil)
	mockStorageGenesisSlot.On("Get").Return(genesisSlot, nil)

	result, err := target.CurrentEpochStart()

	assert.NoError(t, err)
	assert.Equal(t, sc.U64(500), result)

	mockStorageEpochIndex.AssertCalled(t, "Get")
	mockStorageGenesisSlot.AssertCalled(t, "Get")
}

func Test_CurrentEpoch(t *testing.T) {
	target := setupModule()

	mockStorageEpochIndex.On("Get").Return(epochIndex, nil)
	mockStorageGenesisSlot.On("Get").Return(genesisSlot, nil)
	mockStorageAuthorities.On("Get").Return(authorities, nil)
	mockStorageRandomness.On("Get").Return(randomness, nil)
	mockStorageEpochConfig.On("Get").Return(epochConfig, nil)

	epoch, err := target.CurrentEpoch()

	expectedEpoch := babetypes.Epoch{
		EpochIndex:  epochIndex,
		StartSlot:   sc.U64(500),
		Duration:    epochDuration,
		Authorities: authorities,
		Randomness:  randomness,
		Config:      epochConfig,
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedEpoch, epoch)

	mockStorageEpochIndex.AssertCalled(t, "Get")
	mockStorageGenesisSlot.AssertCalled(t, "Get")
	mockStorageAuthorities.AssertCalled(t, "Get")
	mockStorageRandomness.AssertCalled(t, "Get")
	mockStorageEpochConfig.AssertCalled(t, "Get")
}

func Test_NextEpoch(t *testing.T) {
	target := setupModule()

	mockStorageEpochIndex.On("Get").Return(epochIndex, nil)
	mockStorageGenesisSlot.On("Get").Return(genesisSlot, nil)
	mockStorageNextAuthorities.On("Get").Return(nextAuthorities, nil)
	mockStorageNextRandomness.On("Get").Return(nextRandomness, nil)
	mockStorageNextEpochConfig.On("Get").Return(epochConfig, nil)

	epoch, err := target.NextEpoch()

	expectedEpoch := babetypes.Epoch{
		EpochIndex:  epochIndex + 1,
		StartSlot:   sc.U64(700),
		Duration:    epochDuration,
		Authorities: nextAuthorities,
		Randomness:  nextRandomness,
		Config:      epochConfig,
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedEpoch, epoch)

	mockStorageEpochIndex.AssertCalled(t, "Get")
	mockStorageGenesisSlot.AssertCalled(t, "Get")
	mockStorageNextAuthorities.AssertCalled(t, "Get")
	mockStorageNextRandomness.AssertCalled(t, "Get")
	mockStorageNextEpochConfig.AssertCalled(t, "Get")
}

func Test_Module_initializeGenesisAuthorities(t *testing.T) {
	target := setupModule()

	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](nil), nil)
	mockStorageAuthorities.On("Put", authorities).Return(nil)
	mockStorageNextAuthorities.On("Put", authorities).Return(nil)

	err := target.initializeGenesisAuthorities(authorities)

	assert.NoError(t, err)

	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertCalled(t, "Put", authorities)
}

func Test_Module_initializeGenesisAuthorities_Already_Initialized(t *testing.T) {
	target := setupModule()

	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](sc.U64(1)), nil)
	mockStorageAuthorities.On("Put", authorities).Return(nil)
	mockStorageNextAuthorities.On("Put", authorities).Return(nil)

	err := target.initializeGenesisAuthorities(authorities)

	assert.Equal(t, errAuthoritiesAlreadyInitialized, err)

	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertNotCalled(t, "Put", authorities)
}
