package session

import (
	"bytes"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	sessiontypes "github.com/LimeChain/gosemble/primitives/session"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	functionSetKeys = iota
	functionPurgeKeys
)

const (
	name = sc.Str("Session")
)

var (
	dispatchErrNotFound = primitives.NewDispatchErrorOther("not found")
)

type Module interface {
	primitives.Module
	types.InherentProvider

	CurrentIndex() (sc.U32, error)
	Validators() (sc.Sequence[primitives.AccountId], error)
	IsDisabled(index sc.U32) (bool, error)
	DecodeKeys(buffer *bytes.Buffer) (sc.FixedSequence[primitives.Sr25519PublicKey], error)

	AppendHandlers(module sessiontypes.OneSessionHandler)
}

type module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index        sc.U8
	config       Config
	functions    map[sc.U8]primitives.Call
	mdGenerator  *primitives.MetadataTypeGenerator
	storage      *storage
	sessionEnder ShouldEndSession
	systemModule system.Module
	handler      Handler
	manager      Manager
	logger       log.Logger
}

func New(index sc.U8, config Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.Logger) Module {
	functions := make(map[sc.U8]primitives.Call)

	module := module{
		index:        index,
		config:       config,
		functions:    functions,
		systemModule: config.Module,
		sessionEnder: config.SessionEnder,
		mdGenerator:  mdGenerator,
		handler:      config.Handler,
		manager:      config.Manager,
		logger:       logger,
	}
	module.storage = newStorage(module)

	functions[functionSetKeys] = newCallSetKeys(index, functionSetKeys, config.DbWeight, module, config.Handler)
	functions[functionPurgeKeys] = newCallPurgeKeys(index, functionPurgeKeys, config.DbWeight, module)
	module.functions = functions

	return module
}

func (m module) CurrentIndex() (sc.U32, error) {
	return m.storage.CurrentIndex.Get()
}

func (m module) Validators() (sc.Sequence[primitives.AccountId], error) {
	return m.storage.Validators.Get()
}

func (m module) name() sc.Str { return name }

func (m module) GetIndex() sc.U8 { return m.index }

func (m module) Functions() map[sc.U8]primitives.Call { return m.functions }

func (m module) PreDispatch(_ primitives.Call) (sc.Empty, error) { return sc.Empty{}, nil }

func (m module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

func (m module) OnInitialize(n sc.U64) (primitives.Weight, error) {
	if m.sessionEnder.ShouldEndSession(n) {
		return m.config.BlockWeights.MaxBlock, m.rotateSession()
	}

	return primitives.WeightZero(), nil
}

// rotateSession moves on to the next session.
// Registers the new validators set with its corresponding session keys.
func (m module) rotateSession() error {
	sessionIndex, err := m.storage.CurrentIndex.Get()
	if err != nil {
		return err
	}

	m.logger.Tracef("rotating session: [%d]", sessionIndex)

	changed, err := m.storage.QueueChanged.Get()
	if err != nil {
		return err
	}

	// Inform that the session is ending.
	m.handler.OnBeforeSessionEnding()
	m.manager.EndSession(sessionIndex)

	sessionKeys, err := m.storage.QueuedKeys.Get()
	if err != nil {
		return err
	}

	validators := getValidatorsFromQueuedKeys(sessionKeys)
	m.storage.Validators.Put(validators)

	if changed {
		m.storage.DisabledValidators.Clear()
	}

	sessionIndex += 1
	m.storage.CurrentIndex.Put(sessionIndex)

	m.manager.StartSession(sessionIndex)

	maybeNewValidators := m.manager.NewSession(sessionIndex + 1)

	nextValidators, nextIdentitiesChanged, err := m.nextValidators(maybeNewValidators)
	if err != nil {
		return err
	}

	queuedKeys, nextChanged, err := m.queueNextValidators(sessionKeys, nextValidators, nextIdentitiesChanged)
	if err != nil {
		return err
	}

	m.storage.QueuedKeys.Put(queuedKeys)
	m.storage.QueueChanged.Put(sc.Bool(nextChanged))

	m.systemModule.DepositEvent(newEventNewSession(m.index, sessionIndex))

	return m.handler.OnNewSession(bool(changed), sessionKeys, queuedKeys)
}

func (m module) DecodeKeys(buffer *bytes.Buffer) (sc.FixedSequence[primitives.Sr25519PublicKey], error) {
	return m.handler.DecodeKeys(buffer)
}

func (m module) AppendHandlers(module sessiontypes.OneSessionHandler) {
	m.handler.AppendHandlers(module)
}

// DoSetKeys performs the `set_key` operation, checking for duplicates.
func (m module) DoSetKeys(who primitives.AccountId, sessionKeys sc.Sequence[primitives.SessionKey]) error {
	canIncConsumer, err := m.systemModule.CanIncConsumer(who)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	if !canIncConsumer {
		return NewDispatchErrorNoAccount(m.index)
	}

	oldKeys, err := m.innerSetKeys(who, sessionKeys)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	if oldKeys.HasValue && len(oldKeys.Value) > 0 {
		return nil
	}

	err = m.systemModule.IncConsumers(who)
	if err != nil {
		return err
	}
	m.logger.Debug("inc_consumers returned true;")

	return nil
}

func (m module) DoPurgeKeys(who primitives.AccountId) error {
	bytesNextKeys, err := m.storage.NextKeys.TakeBytes(who)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	if bytesNextKeys == nil {
		return NewDispatchErrorNoKeys(m.index)
	}

	nextKeys, err := m.handler.DecodeKeys(bytes.NewBuffer(bytesNextKeys))
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	keyTypeIds := m.handler.KeyTypeIds()

	keys, err := toSessionKeys(keyTypeIds, nextKeys)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	for _, keyTypeId := range keyTypeIds {
		key, found := getKey(keyTypeId, keys)
		if !found {
			return dispatchErrNotFound
		}

		m.storage.KeyOwner.Remove(key)
	}

	err = m.systemModule.DecConsumers(who)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	return nil
}

func (m module) IsDisabled(index sc.U32) (bool, error) {
	disabledValidators, err := m.storage.DisabledValidators.Get()
	if err != nil {
		return false, err
	}

	for _, disabledValidator := range disabledValidators {
		if disabledValidator == index {
			return true, nil
		}
	}

	return false, nil
}

func (m module) StorageDisabledValidators() (sc.Sequence[sc.U32], error) {
	return m.storage.DisabledValidators.Get()
}

// innerSetKeys sets the keys and checks for duplicates.
func (m module) innerSetKeys(who primitives.AccountId, sessionKeys sc.Sequence[primitives.SessionKey]) (sc.Option[sc.Sequence[primitives.SessionKey]], error) {
	oldKeys, err := m.storage.NextKeys.Get(who)
	if err != nil {
		return sc.Option[sc.Sequence[primitives.SessionKey]]{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	keyTypeIds := m.handler.KeyTypeIds()

	for _, keyTypeId := range keyTypeIds {
		key, found := getKey(keyTypeId, sessionKeys)
		if !found {
			return sc.Option[sc.Sequence[primitives.SessionKey]]{}, dispatchErrNotFound
		}

		keyOwner, err := m.storage.KeyOwner.Get(key)
		if err != nil {
			return sc.Option[sc.Sequence[primitives.SessionKey]]{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		}

		if reflect.DeepEqual(keyOwner, who) {
			return sc.Option[sc.Sequence[primitives.SessionKey]]{}, NewDispatchErrorDuplicatedKey(m.index)
		}
	}

	oldSessionKeys, err := toSessionKeys(keyTypeIds, oldKeys)
	if err != nil {
		return sc.Option[sc.Sequence[primitives.SessionKey]]{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	for _, keyTypeId := range keyTypeIds {
		newKey, found := getKey(keyTypeId, sessionKeys)
		if !found {
			return sc.Option[sc.Sequence[primitives.SessionKey]]{}, dispatchErrNotFound
		}

		oldKey, found := getKey(keyTypeId, oldSessionKeys)
		if found {
			if reflect.DeepEqual(newKey, oldKey) {
				continue
			}
			m.storage.KeyOwner.Remove(oldKey)
		}

		m.storage.KeyOwner.Put(newKey, who)
	}

	nextKeys := toPublicKeys(sessionKeys)

	m.storage.NextKeys.Put(who, nextKeys)

	return sc.NewOption[sc.Sequence[primitives.SessionKey]](oldSessionKeys), nil
}

func (m module) Metadata() primitives.MetadataModule {
	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSessionCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSessionCalls, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Session, Runtime>"),
				},
				m.index,
				"Call.Session")),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSessionEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSessionEvent, "frame_session::Event<Runtime>"),
				},
				m.index,
				"Events.System"),
		),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSessionErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesSessionErrors),
				},
				m.index,
				"Errors.Session"),
		),
		Index: m.index,
	}
	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

func (m module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
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
	})
}

func (m module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithPath(
			metadata.TypesSessionEvent,
			"pallet_session pallet Session",
			sc.Sequence[sc.Str]{"pallet_session", "pallet", "Event"},
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
}

func (m module) queueNextValidators(sessionKeys sc.Sequence[queuedKey], nextValidators sc.Sequence[primitives.AccountId], nextIdentitiesChanged bool) (sc.Sequence[queuedKey], bool, error) {
	var result sc.Sequence[queuedKey]

	for i, nextValidator := range nextValidators {
		nextKeys, err := m.storage.NextKeys.Get(nextValidator)
		if err != nil {
			return nil, false, err
		}
		if nextIdentitiesChanged {
			continue
		}
		sessionKey := sessionKeys[i]

		if !reflect.DeepEqual(sessionKey.Keys, nextKeys) {
			nextIdentitiesChanged = true
		}

		keys, err := toSessionKeys(m.handler.KeyTypeIds(), nextKeys)
		if err != nil {
			return nil, false, err
		}

		result = append(result, queuedKey{
			Validator: nextValidator,
			Keys:      keys,
		})
	}

	return result, nextIdentitiesChanged, nil
}

func (m module) nextValidators(maybeNewValidators sc.Option[sc.Sequence[primitives.AccountId]]) (sc.Sequence[primitives.AccountId], bool, error) {
	if maybeNewValidators.HasValue {
		return maybeNewValidators.Value, true, nil
	}

	validators, err := m.storage.Validators.Get()
	if err != nil {
		return nil, false, err
	}
	return validators, false, nil

}

func getKey(keyTypeId sc.FixedSequence[sc.U8], sessionKeys sc.Sequence[primitives.SessionKey]) (primitives.SessionKey, bool) {
	for _, sessionKey := range sessionKeys {
		if reflect.DeepEqual(sessionKey.TypeId, keyTypeId) {
			return sessionKey, true
		}
	}

	return primitives.SessionKey{}, false
}

func getValidatorsFromQueuedKeys(queuedKeys sc.Sequence[queuedKey]) sc.Sequence[primitives.AccountId] {
	var result sc.Sequence[primitives.AccountId]

	for _, queuedKey := range queuedKeys {
		result = append(result, queuedKey.Validator)
	}

	return result
}
