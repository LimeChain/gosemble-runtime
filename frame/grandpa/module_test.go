package grandpa

import (
	"bytes"
	"errors"
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/grandpa"
	grandpatypes "github.com/LimeChain/gosemble/primitives/grandpa"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	moduleId               = sc.U8(1)
	maxAuthorities         = sc.U32(2)
	maxNominators          = sc.U32(2)
	maxSetIdSessionEntries = sc.U64(3)
)

var (
	target                       module
	mockSystemModule             *mocks.SystemModule
	mockSessionModule            *mocks.SessionModule
	mockEquivocationReportSystem *mocks.EquivocationReportSystem
	mockKeyOwnerProofSystem      *mocks.KeyOwnerProofSystem
	mockStorageAuthorities       *mocks.StorageValue[sc.Sequence[primitives.Authority]]
	mockStorageSetIdSession      *mocks.StorageMap[sc.U64, sc.U32]
	mockStorageStalled           *mocks.StorageValue[primitives.Tuple2U64]
	mockStorageCurrentSetId      *mocks.StorageValue[sc.U64]
	mockStoragePendingChange     *mocks.StorageValue[StoredPendingChange]
	mockStorageNextForced        *mocks.StorageValue[sc.U64]
	mockStorageState             *mocks.StorageValue[StoredState]
	logger                       = log.NewLogger()
	mdGenerator                  = primitives.NewMetadataTypeGenerator()
)

var (
	accountId1, _    = types.NewAccountId(constants.OneAddress.FixedSequence...)
	accountId2, _    = types.NewAccountId(constants.TwoAddress.FixedSequence...)
	pubKey1          = primitives.Ed25519PublicKey{FixedSequence: accountId1.FixedSequence}
	pubKey2          = primitives.Ed25519PublicKey{FixedSequence: accountId2.FixedSequence}
	emptyAuthorities = sc.Sequence[primitives.Authority]{}
	authorities      = sc.Sequence[types.Authority]{
		{
			Id:     accountId1,
			Weight: 1,
		},
	}
	validators = sc.Sequence[primitives.Validator]{
		primitives.Validator{
			AccountId:   accountId1,
			AuthorityId: primitives.Sr25519PublicKey{FixedSequence: accountId1.FixedSequence},
		},
	}
	queuedValidators = sc.Sequence[primitives.Validator]{
		primitives.Validator{
			AccountId:   accountId2,
			AuthorityId: primitives.Sr25519PublicKey{FixedSequence: accountId2.FixedSequence},
		},
	}

	currentBlock             = sc.U64(15)
	waitInBlocks             = sc.U64(5)
	medianLastFinalizedBlock = sc.U64(18)
	forcedAtBlock            = medianLastFinalizedBlock

	dbWeight = primitives.RuntimeDbWeight{
		Read:  1,
		Write: 2,
	}

	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
)

func setup() {
	mockSystemModule = new(mocks.SystemModule)
	mockSessionModule = new(mocks.SessionModule)
	mockEquivocationReportSystem = new(mocks.EquivocationReportSystem)
	mockKeyOwnerProofSystem = new(mocks.KeyOwnerProofSystem)
	mockStorageAuthorities = new(mocks.StorageValue[sc.Sequence[primitives.Authority]])
	mockStorageSetIdSession = new(mocks.StorageMap[sc.U64, sc.U32])
	mockStorageStalled = new(mocks.StorageValue[primitives.Tuple2U64])
	mockStorageCurrentSetId = new(mocks.StorageValue[sc.U64])
	mockStoragePendingChange = new(mocks.StorageValue[StoredPendingChange])
	mockStorageNextForced = new(mocks.StorageValue[sc.U64])
	mockStorageState = new(mocks.StorageValue[StoredState])

	config := NewConfig(
		dbWeight,
		primitives.PublicKeyEd25519,
		maxAuthorities,
		maxNominators,
		maxSetIdSessionEntries,
		mockKeyOwnerProofSystem,
		mockEquivocationReportSystem,
		mockSystemModule,
		mockSessionModule,
	)
	target = New(moduleId, config, logger, mdGenerator).(module)

	target.storage.Authorities = mockStorageAuthorities
	target.storage.SetIdSession = mockStorageSetIdSession
	target.storage.Stalled = mockStorageStalled
	target.storage.CurrentSetId = mockStorageCurrentSetId
	target.storage.PendingChange = mockStoragePendingChange
	target.storage.NextForced = mockStorageNextForced
	target.storage.State = mockStorageState
}

func Test_Grandpa_Module_GetIndex(t *testing.T) {
	setup()

	assert.Equal(t, moduleId, target.GetIndex())
}

func Test_Grandpa_Module_Functions(t *testing.T) {
	setup()

	assert.Equal(t, 3, len(target.Functions()))
}

func Test_Grandpa_Module_PreDispatch(t *testing.T) {
	setup()

	result, err := target.PreDispatch(new(mocks.Call))

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Grandpa_Module_ValidateUnsigned(t *testing.T) {
	setup()

	result, err := target.ValidateUnsigned(primitives.NewTransactionSourceLocal(), new(mocks.Call))

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Grandpa_Module_KeyType(t *testing.T) {
	setup()
	assert.Equal(t, primitives.PublicKeyEd25519, target.KeyType())
}

func Test_Grandpa_Module_KeyTypeId(t *testing.T) {
	setup()
	assert.Equal(t, KeyTypeId, target.KeyTypeId())
}

func Test_Grandpa_Module_DecodeKey(t *testing.T) {
	setup()

	buffer := bytes.NewBuffer(pubKey1.Bytes())

	result, err := target.DecodeKey(buffer)

	assert.NoError(t, err)
	assert.Equal(t, pubKey1, result)
}

func Test_Grandpa_Module_DecodeKey_Fails(t *testing.T) {
	setup()

	buffer := bytes.NewBuffer(pubKey1.Bytes()[:1])

	_, err := target.DecodeKey(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_Grandpa_Module_OnGenesisSession(t *testing.T) {
	setup()

	mockStorageAuthorities.On("Get").Return(emptyAuthorities, nil)
	mockStorageAuthorities.On("Put", authorities).Return(nil)
	mockStorageSetIdSession.On("Put", sc.U64(0), sc.U32(0)).Return(nil)

	err := target.OnGenesisSession(validators)

	assert.NoError(t, err)
	mockStorageAuthorities.AssertCalled(t, "Get")
	mockStorageAuthorities.AssertCalled(t, "Put", authorities)
	mockStorageSetIdSession.AssertCalled(t, "Put", sc.U64(0), sc.U32(0))
}

func Test_Grandpa_Module_OnGenesisSession_Decode_Error(t *testing.T) {
	setup()

	mockStorageAuthorities.On("Get").Return(emptyAuthorities, errors.New("decode error"))

	err := target.OnGenesisSession(validators)

	assert.Error(t, err)
	mockStorageAuthorities.AssertCalled(t, "Get")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageSetIdSession.AssertNotCalled(t, "Put", sc.U64(0), sc.U32(0))
}

func Test_Grandpa_Module_OnGenesisSession_Initialized_Authorities_Error(t *testing.T) {
	setup()

	mockStorageAuthorities.On("Get").Return(authorities, nil)

	err := target.OnGenesisSession(validators)

	assert.Equal(t, errAuthoritiesAlreadyInitialized, err)
	mockStorageAuthorities.AssertCalled(t, "Get")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageSetIdSession.AssertNotCalled(t, "Put", sc.U64(0), sc.U32(0))
}

func Test_Grandpa_Module_OnGenesisSession_MaxAuthorities_Error(t *testing.T) {
	setup()

	mockStorageAuthorities.On("Get").Return(authorities, nil)

	err := target.OnGenesisSession(validators)

	assert.Equal(t, errAuthoritiesAlreadyInitialized, err)
	mockStorageAuthorities.AssertCalled(t, "Get")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageSetIdSession.AssertNotCalled(t, "Put", sc.U64(0), sc.U32(0))
}

func Test_Grandpa_Module_OnNewSession_NoChanges(t *testing.T) {
	setup()

	currentSetId := sc.U64(0)
	currentSessionIndex := sc.U32(0)

	mockStorageStalled.On("Exists").Return(false, nil)
	mockStorageCurrentSetId.On("Get").Return(currentSetId, nil)
	mockSessionModule.On("CurrentIndex").Return(currentSessionIndex, nil)
	mockStorageSetIdSession.On("Put", currentSetId, currentSessionIndex).Return(nil)

	err := target.OnNewSession(false, validators, queuedValidators)

	assert.Nil(t, err)
}

func Test_Grandpa_Module_OnNewSession_Stalled(t *testing.T) {
	setup()

	currentSetId := sc.U64(4)
	nextSetId := currentSetId + 1
	currentSessionIndex := sc.U32(maxSetIdSessionEntries)

	nextForced := sc.U64(14)
	stalled := primitives.Tuple2U64{First: waitInBlocks, Second: forcedAtBlock}
	pendingChange := StoredPendingChange{
		Delay:           waitInBlocks,
		ScheduledAt:     currentBlock,
		NextAuthorities: authorities,
		Forced:          sc.NewOption[sc.U64](forcedAtBlock),
	}

	mockMutateFnSetId := mock.AnythingOfType("func(*goscale.U64) (goscale.U64, error)")

	mockStorageStalled.On("Exists").Return(true, nil)
	mockStorageStalled.On("TakeBytes").Return(stalled.Bytes(), nil)
	mockStoragePendingChange.On("Exists").Return(false, nil)
	mockSystemModule.On("StorageBlockNumber").Return(currentBlock, nil)
	mockStorageNextForced.On("Get").Return(nextForced, nil)
	mockStorageNextForced.On("Put", currentBlock+waitInBlocks*2).Return(nil)
	mockStoragePendingChange.On("Put", pendingChange).Return(nil)
	mockStorageCurrentSetId.On("Mutate", mockMutateFnSetId).Return(nextSetId, nil)
	mockStorageSetIdSession.On("Remove", nextSetId-maxSetIdSessionEntries)
	mockStorageCurrentSetId.On("Get").Return(nextSetId, nil)
	mockSessionModule.On("CurrentIndex").Return(currentSessionIndex, nil)
	mockStorageSetIdSession.On("Put", nextSetId, currentSessionIndex).Return(nil)

	err := target.OnNewSession(false, validators, queuedValidators)

	assert.NoError(t, err)
	mockStorageStalled.AssertCalled(t, "Exists")
	mockStorageStalled.AssertCalled(t, "TakeBytes")
	mockStoragePendingChange.AssertCalled(t, "Exists")
	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
	mockStorageNextForced.AssertCalled(t, "Get")
	mockStorageNextForced.AssertCalled(t, "Put", currentBlock+waitInBlocks*2)
	mockStoragePendingChange.AssertCalled(t, "Put", pendingChange)
	mockStorageCurrentSetId.AssertCalled(t, "Mutate", mockMutateFnSetId)
	mockStorageSetIdSession.AssertCalled(t, "Remove", nextSetId-maxSetIdSessionEntries)
	mockSessionModule.AssertCalled(t, "CurrentIndex")
	mockStorageSetIdSession.AssertCalled(t, "Put", nextSetId, currentSessionIndex)
}

func Test_Grandpa_Module_OnNewSession_Stalled_And_Pending(t *testing.T) {
	setup()

	currentSetId := sc.U64(0)
	currentSessionIndex := sc.U32(0)
	stalled := primitives.Tuple2U64{First: 10, Second: 20}

	mockStorageStalled.On("Exists").Return(true, nil)
	mockStorageStalled.On("TakeBytes").Return(stalled.Bytes(), nil)
	mockStoragePendingChange.On("Exists").Return(true, nil)
	mockStorageCurrentSetId.On("Get").Return(currentSetId, nil)
	mockSessionModule.On("CurrentIndex").Return(currentSessionIndex, nil)
	mockStorageSetIdSession.On("Put", currentSetId, currentSessionIndex).Return(nil)

	err := target.OnNewSession(false, validators, queuedValidators)

	assert.NoError(t, err)
	mockStorageStalled.AssertCalled(t, "Exists")
	mockStorageStalled.AssertCalled(t, "TakeBytes")
	mockStoragePendingChange.AssertCalled(t, "Exists")
	mockStorageCurrentSetId.AssertCalled(t, "Get")
	mockSessionModule.AssertCalled(t, "CurrentIndex")
	mockStorageSetIdSession.AssertCalled(t, "Put", currentSetId, currentSessionIndex)
}

func Test_Grandpa_Module_OnDisabled(t *testing.T) {
	setup()

	validatorIndex := sc.U32(10)

	digest := primitives.NewDigestItemConsensusMessage(
		sc.BytesToFixedSequenceU8([]byte{'F', 'R', 'N', 'K'}),
		sc.BytesToSequenceU8(grandpatypes.NewConsensusLogOnDisabled(sc.U64(validatorIndex)).Bytes()),
	)
	mockSystemModule.On("DepositLog", digest).Return(nil)

	target.OnDisabled(validatorIndex)

	mockSystemModule.AssertCalled(t, "DepositLog", digest)
}

func Test_Grandpa_Module_OnFinalize_With_Delay_After_Current_Block(t *testing.T) {
	setup()

	pendingChange := StoredPendingChange{
		Delay:           waitInBlocks,
		ScheduledAt:     currentBlock,
		NextAuthorities: authorities,
		Forced:          sc.NewOption[sc.U64](forcedAtBlock),
	}
	pendingChangeBytes := sc.BytesToSequenceU8(pendingChange.Bytes())

	mockStoragePendingChange.On("GetBytes").Return(sc.NewOption[sc.Sequence[sc.U8]](pendingChangeBytes), nil)
	mockStorageAuthorities.On("Put", authorities).Return(nil)
	mockSystemModule.On("DepositEvent", newEventNewAuthorities(moduleId, authorities)).Return(nil)
	mockStoragePendingChange.On("Clear").Return(nil)
	mockStorageState.On("Get").Return(NewStoredStateLive(), nil)

	err := target.OnFinalize(currentBlock + waitInBlocks)

	assert.Nil(t, err)

	mockStoragePendingChange.AssertCalled(t, "GetBytes")
	mockStorageAuthorities.AssertCalled(t, "Put", authorities)
	mockSystemModule.AssertCalled(t, "DepositEvent", newEventNewAuthorities(moduleId, authorities))
	mockStoragePendingChange.AssertCalled(t, "Clear")
	mockStorageState.AssertCalled(t, "Get")
}

func Test_Grandpa_Module_OnFinalize_At_Current_Block(t *testing.T) {
	setup()

	pendingChange := StoredPendingChange{
		Delay:           waitInBlocks,
		ScheduledAt:     currentBlock,
		NextAuthorities: authorities,
		Forced:          sc.NewOption[sc.U64](forcedAtBlock),
	}
	pendingChangeBytes := sc.BytesToSequenceU8(pendingChange.Bytes())

	scheduledChange := grandpa.ScheduledChange{
		NextAuthorities: authorities,
		Delay:           pendingChange.Delay,
	}
	digest := primitives.NewDigestItemConsensusMessage(
		sc.BytesToFixedSequenceU8([]byte{'F', 'R', 'N', 'K'}),
		sc.BytesToSequenceU8(grandpatypes.NewConsensusLogForcedChange(forcedAtBlock, scheduledChange).Bytes()),
	)

	mockStoragePendingChange.On("GetBytes").Return(sc.NewOption[sc.Sequence[sc.U8]](pendingChangeBytes), nil)
	mockSystemModule.On("DepositLog", digest).Return(nil)
	mockStorageState.On("Get").Return(NewStoredStateLive(), nil)

	err := target.OnFinalize(currentBlock)

	assert.Nil(t, err)

	mockStoragePendingChange.AssertCalled(t, "GetBytes")
	mockSystemModule.AssertCalled(t, "DepositLog", digest)
	mockStorageState.AssertCalled(t, "Get")
}

func Test_Grandpa_Module_StorageSetId(t *testing.T) {
	setup()

	currentSetId := sc.U64(1)

	mockStorageCurrentSetId.On("Get").Return(currentSetId, nil)

	result, err := target.StorageSetId()

	assert.NoError(t, err)
	assert.Equal(t, currentSetId, result)
	mockStorageCurrentSetId.AssertCalled(t, "Get")
}

func Test_Grandpa_Module_Metadata(t *testing.T) {
	setup()

	expectMetadataTypes := sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithParams(
			metadata.TypesGrandpaErrors,
			"The `Error` enum of this pallet.",
			sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Error"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant("PauseFailed", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorPauseFailed, ""),
					primitives.NewMetadataDefinitionVariant("ResumeFailed", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorResumeFailed, ""),
					primitives.NewMetadataDefinitionVariant("ChangePending", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorChangePending, ""),
					primitives.NewMetadataDefinitionVariant("TooSoon", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorTooSoon, ""),
					primitives.NewMetadataDefinitionVariant("InvalidKeyOwnershipProof", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorInvalidKeyOwnershipProof, ""),
					primitives.NewMetadataDefinitionVariant("InvalidEquivocationProof", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorInvalidEquivocationProof, ""),
					primitives.NewMetadataDefinitionVariant("DuplicateOffenceReport", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorDuplicateOffenceReport, ""),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
			},
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesGrandpaAppPublic,
			"sp_consensus_grandpa app Public",
			sc.Sequence[sc.Str]{"sp_consensus_grandpa", "app", "Public"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesEd25519PubKey),
				},
			),
		),

		primitives.NewMetadataType(
			metadata.TypesTupleGrandpaAppPublicU64,
			"(GrandpaAppPublic, U64)",
			primitives.NewMetadataTypeDefinitionTuple(
				sc.Sequence[sc.Compact]{
					sc.ToCompact(metadata.TypesGrandpaAppPublic), sc.ToCompact(metadata.PrimitiveTypesU64),
				},
			),
		),

		primitives.NewMetadataType(
			metadata.TypesSequenceTupleGrandpaAppPublic,
			"[]byte (GrandpaAppPublic, U64)",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleGrandpaAppPublicU64)),
		),

		primitives.NewMetadataTypeWithParams(
			metadata.TypesGrandpaStoredPendingChange,
			"StoredPendingChange",
			sc.Sequence[sc.Str]{"pallet_grandpa", "StoredPendingChange"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "scheduled_at", "u64"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "delay", "u64"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesBoundedVecAuthority, "next_authorities", "BoundedAuthorityList<Limit>"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionU64, "forced", "Option<u64>"),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU64, "N"),
				primitives.NewMetadataEmptyTypeParameter("Limit"),
			},
		),

		primitives.NewMetadataTypeWithParams(
			metadata.TypesGrandpaStoredState,
			"StoredState",
			sc.Sequence[sc.Str]{"pallet_grandpa", "StoredState"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant("Live", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, StoredStateLive, ""),
					primitives.NewMetadataDefinitionVariant("PendingPause", sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "scheduled_at", "u64"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "delay", "u64"),
					}, StoredStatePendingPause, ""),
					primitives.NewMetadataDefinitionVariant("Paused", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, StoredStatePaused, ""),
					primitives.NewMetadataDefinitionVariant("PendingResume", sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "scheduled_at", "u64"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "delay", "u64"),
					}, StoredStatePendingResume, ""),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU64, "N"),
			},
		),

		primitives.NewMetadataTypeWithParam(
			metadata.TypesGrandpaCalls,
			"Grandpa calls",
			sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"note_stalled",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU64, "delay"),
							primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU64, "best_finalized_block_number"),
						},
						functionNoteStalledIndex,
						"Note that the current authority set of the GRANDPA finality gadget has stalled. This will trigger a forced authority set change at the beginning of the next session, to be enacted `delay` blocks after that. The `delay` should be high enough to safely assume that the block signalling the forced change will not be re-orged e.g. 1000 blocks. The block production rate (which may be slowed down because of finality lagging) should be taken into account when choosing the `delay`. The GRANDPA voters based on the new authority will start voting on top of `best_finalized_block_number` for new finalized blocks. `best_finalized_block_number` should be the highest of the latest finalized block of all validators of the new authority set. Only callable by root.",
					),
				},
			),
			primitives.NewMetadataEmptyTypeParameter("T"),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesGrandpaEvent,
			"pallet_grandpa pallet Event",
			sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Event"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"NewAuthorities",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceAuthority, "authority_set", "AuthorityList"),
						},
						EventNewAuthorities,
						"New authority set has been applied.",
					),

					primitives.NewMetadataDefinitionVariant(
						"Paused",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventPaused,
						"Current authority set has been paused.",
					),

					primitives.NewMetadataDefinitionVariant(
						"Resumed",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventResumed,
						"Current authority set has been resumed.",
					),
				},
			),
		),
	}

	moduleV14 := primitives.MetadataModuleV14{
		Name: "Grandpa",
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: "Grandpa",
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"Authorities",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBoundedVecAuthority)),
					"The current list of authorities.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"CurrentSetId",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU64)),
					"The number of changes (both in terms of keys and underlying economic responsibilities) in the \"set\" of Grandpa validators from genesis.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"Stalled",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesTuple2U64)),
					"`true` if we are currently stalled.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"PendingChange",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesGrandpaStoredPendingChange)),
					"Pending change: (signaled at, scheduled change).",
				),
				primitives.NewMetadataModuleStorageEntry(
					"State",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesGrandpaStoredState)),
					"State of the current authority set.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"SetIdSession",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
						sc.ToCompact(metadata.PrimitiveTypesU64),
						sc.ToCompact(metadata.PrimitiveTypesU32),
					),
					"A mapping from grandpa set ID to the index of the *most recent* session for which its members were responsible. This is only used for validating equivocation proofs. An equivocation proof must contains a key-ownership proof for a given session, therefore we need a way to tie together sessions and GRANDPA set ids, i.e. we need to validate that a validator was the owner of a given key on a given session, and what the active set ID was during that session. TWOX-NOTE: `SetId` is not under user control.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"NextForced",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU64)),
					"Next block number where we can force a change.",
				),
			},
		}),
		Call: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesGrandpaCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				"Grandpa",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(
						metadata.TypesGrandpaCalls,
						"self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Grandpa, Runtime>",
					),
				},
				moduleId,
				"Call.Grandpa",
			),
		),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesGrandpaEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				"Grandpa",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesGrandpaEvent, "pallet_grandpa::Event<Runtime>"),
				},
				moduleId,
				"Events.Grandpa",
			),
		),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{
			primitives.NewMetadataModuleConstant(
				"MaxAuthorities",
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.BytesToSequenceU8(maxAuthorities.Bytes()),
				"Max Authorities in use.",
			),
			primitives.NewMetadataModuleConstant(
				"MaxNominators",
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.BytesToSequenceU8(maxNominators.Bytes()),
				"The maximum number of nominators for each validator.",
			),
			primitives.NewMetadataModuleConstant(
				"MaxSetIdSessionEntries",
				sc.ToCompact(metadata.PrimitiveTypesU64),
				sc.BytesToSequenceU8(maxSetIdSessionEntries.Bytes()),
				"The maximum number of entries to keep in the set id to session index mapping. Since the `SetIdSession` map is only used for validating equivocations this value should relate to the bonding duration of whatever staking system is being used (if any). If equivocation handling is not enabled then this value can be zero.",
			),
		},
		Error: sc.NewOption[sc.Compact](nil),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				"Grandpa",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesGrandpaErrors),
				},
				moduleId,
				"Errors.Grandpa",
			),
		),
		Index: moduleId,
	}

	expectMetadataModule := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: moduleV14,
	}

	metadataModule := target.Metadata()
	metadataTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectMetadataTypes, metadataTypes)
	assert.Equal(t, expectMetadataModule, metadataModule)
}
