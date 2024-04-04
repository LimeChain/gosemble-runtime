package session

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	moduleId = 2
)

var (
	blockNumber  = sc.U64(5)
	blockWeights = primitives.BlockWeights{
		MaxBlock: primitives.Weight{
			RefTime:   3,
			ProofSize: 4,
		},
	}
	dbWeight = primitives.RuntimeDbWeight{
		Read:  1,
		Write: 2,
	}

	key              = sc.Sequence[sc.U8]{'k', 'e', 'y'}
	sr25519PublicKey = primitives.Sr25519PublicKey{FixedSequence: sc.FixedSequence[sc.U8]{'k', 'e', 'y'}}
	sessionKey       = primitives.SessionKey{
		Key:    key,
		TypeId: sc.FixedSequence[sc.U8]{'t', 'e', 's', 't'},
	}
	sessionKeys = sc.Sequence[primitives.SessionKey]{
		sessionKey,
	}
	keyTypeIds = sc.Sequence[sc.FixedSequence[sc.U8]]{
		sc.FixedSequence[sc.U8]{'t', 'e', 's', 't'},
	}
	nextKeys = sc.FixedSequence[primitives.Sr25519PublicKey]{
		sr25519PublicKey,
	}
)

var (
	mdGenerator                           = primitives.NewMetadataTypeGenerator()
	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
)

var (
	mockSystemModule     *mocks.SystemModule
	mockShouldEndSession *mocks.ShouldEndSession
	mockSessionHandler   *MockSessionHandler

	mockStorageValidators         *mocks.StorageValue[sc.Sequence[primitives.AccountId]]
	mockStorageCurrentIndex       *mocks.StorageValue[sc.U32]
	mockStorageQueueChanged       *mocks.StorageValue[sc.Bool]
	mockStorageQueuedKeys         *mocks.StorageValue[sc.Sequence[queuedKey]]
	mockStorageDisabledValidators *mocks.StorageValue[sc.Sequence[sc.U32]]
	mockStorageNextKeys           *mocks.StorageMap[primitives.AccountId, sc.FixedSequence[primitives.Sr25519PublicKey]]
	mockStorageKeyOwner           *mocks.StorageMap[primitives.SessionKey, primitives.AccountId]
)

func Test_Module_GetIndex(t *testing.T) {
	target := setupModule()

	assert.Equal(t, sc.U8(moduleId), target.GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	target := setupModule()
	functions := target.Functions()

	assert.Equal(t, 2, len(functions))
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

func Test_Module_OnInitialize(t *testing.T) {
	target := setupModule()

	setKeys := sc.Sequence[queuedKey]{
		{
			Validator: constants.OneAccountId,
			Keys:      sc.Sequence[primitives.SessionKey](nil),
		},
	}

	mockShouldEndSession.On("ShouldEndSession", blockNumber).Return(true)
	mockStorageCurrentIndex.On("Get").Return(sessionIndex, nil)
	mockStorageQueueChanged.On("Get").Return(sc.Bool(true), nil)
	mockSessionHandler.On("OnBeforeSessionEnding").Return()

	mockStorageQueuedKeys.On("Get").Return(queuedKeys, nil)
	mockStorageValidators.On("Put", keys).Return()
	mockStorageDisabledValidators.On("Clear").Return()
	mockStorageCurrentIndex.On("Put", sessionIndex+1).Return()
	mockStorageValidators.On("Get").Return(keys, nil)
	mockStorageNextKeys.On("Get", constants.OneAccountId).Return(sc.FixedSequence[primitives.Sr25519PublicKey]{}, nil)
	mockSessionHandler.On("KeyTypeIds").Return(sc.Sequence[sc.FixedSequence[sc.U8]]{})
	mockStorageQueuedKeys.On("Put", setKeys).Return()
	mockStorageQueueChanged.On("Put", sc.Bool(true)).Return()
	mockSystemModule.On("DepositEvent", newEventNewSession(moduleId, sessionIndex+1)).Return()
	mockSessionHandler.On("OnNewSession", true, queuedKeys, setKeys).Return(nil)

	weight, err := target.OnInitialize(blockNumber)
	assert.Nil(t, err)
	assert.Equal(t, blockWeights.MaxBlock, weight)

	mockShouldEndSession.AssertCalled(t, "ShouldEndSession", blockNumber)
	mockStorageCurrentIndex.AssertCalled(t, "Get")
	mockStorageQueueChanged.AssertCalled(t, "Get")
	mockSessionHandler.AssertCalled(t, "OnBeforeSessionEnding")
	mockStorageQueuedKeys.AssertCalled(t, "Get")
	mockStorageValidators.AssertCalled(t, "Put", keys)
	mockStorageDisabledValidators.AssertCalled(t, "Clear")
	mockStorageCurrentIndex.AssertCalled(t, "Put", sessionIndex+1)
	mockStorageValidators.AssertCalled(t, "Get")
	mockStorageNextKeys.AssertCalled(t, "Get", constants.OneAccountId)
	mockSessionHandler.AssertCalled(t, "KeyTypeIds")
	mockStorageQueuedKeys.AssertCalled(t, "Put", setKeys)
	mockStorageQueueChanged.AssertCalled(t, "Put", sc.Bool(true))
	mockSystemModule.AssertCalled(t, "DepositEvent", newEventNewSession(moduleId, sessionIndex+1))
	mockSessionHandler.AssertCalled(t, "OnNewSession", true, queuedKeys, setKeys)
}

func Test_Module_OnInitialize_NotEndingSession(t *testing.T) {
	target := setupModule()

	mockShouldEndSession.On("ShouldEndSession", blockNumber).Return(false)

	result, err := target.OnInitialize(blockNumber)
	assert.Nil(t, err)
	assert.Equal(t, primitives.WeightZero(), result)

	mockShouldEndSession.AssertCalled(t, "ShouldEndSession", blockNumber)
}

func Test_Module_DoSetKeys(t *testing.T) {
	target := setupModule()

	mockSystemModule.On("CanIncConsumer", constants.OneAccountId).Return(true, nil)
	mockStorageNextKeys.On("Get", constants.OneAccountId).Return(sc.FixedSequence[primitives.Sr25519PublicKey]{}, nil)
	mockSessionHandler.On("KeyTypeIds").Return(keyTypeIds)
	mockStorageKeyOwner.On("Get", sessionKey).Return(constants.ZeroAccountId, nil)
	mockSystemModule.On("IncConsumers", constants.OneAccountId).Return(nil)
	mockStorageKeyOwner.On("Put", sessionKey, constants.OneAccountId).Return()
	mockStorageNextKeys.On("Put", constants.OneAccountId, nextKeys).Return()

	err := target.DoSetKeys(constants.OneAccountId, sessionKeys)
	assert.Nil(t, err)

	mockSystemModule.AssertCalled(t, "CanIncConsumer", constants.OneAccountId)
	mockStorageNextKeys.AssertCalled(t, "Get", constants.OneAccountId)
	mockSessionHandler.AssertCalled(t, "KeyTypeIds")
	mockStorageKeyOwner.AssertCalled(t, "Get", sessionKey)
	mockSystemModule.AssertCalled(t, "IncConsumers", constants.OneAccountId)
	mockStorageKeyOwner.AssertCalled(t, "Put", sessionKey, constants.OneAccountId)
	mockStorageNextKeys.AssertCalled(t, "Put", constants.OneAccountId, nextKeys)
}

func Test_Module_DoPurgeKeys(t *testing.T) {
	target := setupModule()

	mockStorageNextKeys.On("TakeBytes", constants.OneAccountId).Return(nextKeys.Bytes(), nil)
	mockSessionHandler.On("DecodeKeys", bytes.NewBuffer(nextKeys.Bytes())).Return(nextKeys, nil)
	mockSessionHandler.On("KeyTypeIds").Return(keyTypeIds)
	mockStorageKeyOwner.On("Remove", sessionKey).Return()
	mockSystemModule.On("DecConsumers", constants.OneAccountId).Return(nil)

	err := target.DoPurgeKeys(constants.OneAccountId)
	assert.Nil(t, err)

	mockStorageNextKeys.AssertCalled(t, "TakeBytes", constants.OneAccountId)
	mockSessionHandler.AssertCalled(t, "DecodeKeys", bytes.NewBuffer(nextKeys.Bytes()))
	mockSessionHandler.AssertCalled(t, "KeyTypeIds")
	mockStorageKeyOwner.AssertCalled(t, "Remove", sessionKey)
	mockSystemModule.AssertCalled(t, "DecConsumers", constants.OneAccountId)
}

func Test_Module_Metadata(t *testing.T) {
	target := setupModule()

	dataV14 := primitives.MetadataModuleV14{
		Name: name,
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: name,
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"Validators",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceAddress32)),
					"The current set of validators."),
				primitives.NewMetadataModuleStorageEntry(
					"CurrentIndex",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32),
					),
					"Current index of the session.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"QueueChanged",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesBool),
					),
					"True if the underlying economic identities or weighing behind the validators has changed in the queue validator set",
				),
				primitives.NewMetadataModuleStorageEntry(
					"QueuedKeys",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.TypeSequenceQueuedKey),
					),
					"The queued keys for the next session. When the next session begins, these keys will be used to determine the validator's session keys.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"DisabledValidators",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.TypesSequenceU32)),
					"Indices of disabled validators."),
				primitives.NewMetadataModuleStorageEntry(
					"NextKeys",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
						sc.ToCompact(metadata.TypesAddress32),
						sc.ToCompact(metadata.TypesSessionKey)),
					"The next keys for a validator."),
				primitives.NewMetadataModuleStorageEntry(
					"KeyOwner",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
						sc.ToCompact(metadata.TypesSessionStorageKeyOwner),
						sc.ToCompact(metadata.TypesAddress32),
					),
					"The owner of a key. They key is they `KeyTypeId` + encoded key",
				),
			},
		}),
		Call: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSessionCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSessionCalls, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Session, Runtime>"),
				},
				moduleId,
				"Call.Session")),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSessionEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSessionEvent, "frame_session::Event<Runtime>"),
				},
				moduleId,
				"Events.System"),
		),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSessionErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesSessionErrors),
				},
				moduleId,
				"Errors.Session"),
		),
		Index: moduleId,
	}

	expectMetadata := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}

	expectMetadataTypes := sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithPath(
			metadata.TypesSessionEvent,
			"pallet_session pallet Session",
			sc.Sequence[sc.Str]{"pallet_session", "pallet", "Session"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"NewSession",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "session_index", "u32"),
						},
						EventNewSession,
						"Events.NewSession"),
				}),
		),

		primitives.NewMetadataTypeWithParams(metadata.TypesSessionErrors,
			"pallet_session pallet Error",
			sc.Sequence[sc.Str]{"pallet_session", "pallet", "Error"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"InvalidProof",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorInvalidProof,
						"Invalid ownership proof."),
					primitives.NewMetadataDefinitionVariant(
						"NoAssociatedValidatorId",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorNoAssociatedValidatorId,
						"No associated validator ID for account."),
					primitives.NewMetadataDefinitionVariant(
						"DuplicatedKey",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorDuplicatedKey,
						"Registered duplicate key."),
					primitives.NewMetadataDefinitionVariant(
						"NoKeys",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorNoKeys,
						"No keys are associated with this account."),
					primitives.NewMetadataDefinitionVariant(
						"NoAccount",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorNoAccount,
						"Key setting account is not live, so it's impossible to associate keys."),
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			}),

		primitives.NewMetadataTypeWithPath(metadata.TypesSessionKey,
			"runtime SessionKeys",
			sc.Sequence[sc.Str]{"node_template_runtime", "SessionKeys"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSr25519PubKey, "aura", "<Aura as $crate::BoundToRuntimeAppPublic>::Public"),
			}),
		),

		primitives.NewMetadataType(
			metadata.TypesQueuedKey,
			"<Address32, SessionKey>",
			primitives.NewMetadataTypeDefinitionTuple(
				sc.Sequence[sc.Compact]{
					sc.ToCompact(metadata.TypesAddress32),
					sc.ToCompact(metadata.TypesSessionKey),
				})),
		primitives.NewMetadataType(
			metadata.TypeSequenceQueuedKey,
			"[]{Address32,SessionKey}",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesQueuedKey)),
		),
		primitives.NewMetadataType(
			metadata.TypesSessionStorageKeyOwner,
			"<KeyTypeId, Vec<u8>>",
			primitives.NewMetadataTypeDefinitionTuple(
				sc.Sequence[sc.Compact]{
					sc.ToCompact(metadata.TypesSequenceU8),
					sc.ToCompact(metadata.TypesKeyTypeId),
				},
			),
		),

		primitives.NewMetadataTypeWithParam(metadata.TypesSessionCalls,
			"Session calls",
			sc.Sequence[sc.Str]{"frame_session", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"set_keys",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSessionKey, "keys", "T::Keys"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceU8, "proof", "Vec<u8>"),
						},
						functionSetKeys,
						"Sets the session key(s) of the function caller to `keys`."),
					primitives.NewMetadataDefinitionVariant(
						"purge_keys",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						functionPurgeKeys,
						"Removes any session key(s) of the function caller."),
				}),
			primitives.NewMetadataEmptyTypeParameter("T")),
	}

	metadataModule := target.Metadata()
	metadataTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectMetadataTypes, metadataTypes)
	assert.Equal(t, expectMetadata, metadataModule)
}

func setupModule() Module {
	mockSystemModule = new(mocks.SystemModule)
	mockShouldEndSession = new(mocks.ShouldEndSession)
	mockSessionHandler = new(MockSessionHandler)

	config := NewConfig(dbWeight, blockWeights, mockSystemModule, mockShouldEndSession, mockSessionHandler, DefaultManager{})

	initMockStorage()

	target := New(moduleId, config, mdGenerator, log.NewLogger())
	target.storage.Validators = mockStorageValidators
	target.storage.CurrentIndex = mockStorageCurrentIndex
	target.storage.QueueChanged = mockStorageQueueChanged
	target.storage.QueuedKeys = mockStorageQueuedKeys
	target.storage.DisabledValidators = mockStorageDisabledValidators
	target.storage.NextKeys = mockStorageNextKeys
	target.storage.KeyOwner = mockStorageKeyOwner

	return target
}

func initMockStorage() {
	mockStorageValidators = new(mocks.StorageValue[sc.Sequence[primitives.AccountId]])
	mockStorageCurrentIndex = new(mocks.StorageValue[sc.U32])
	mockStorageQueueChanged = new(mocks.StorageValue[sc.Bool])
	mockStorageQueuedKeys = new(mocks.StorageValue[sc.Sequence[queuedKey]])
	mockStorageDisabledValidators = new(mocks.StorageValue[sc.Sequence[sc.U32]])
	mockStorageNextKeys = new(mocks.StorageMap[primitives.AccountId, sc.FixedSequence[primitives.Sr25519PublicKey]])
	mockStorageKeyOwner = new(mocks.StorageMap[primitives.SessionKey, primitives.AccountId])
}
