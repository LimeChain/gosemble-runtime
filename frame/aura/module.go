package aura

import (
	"bytes"
	"errors"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/log"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/hooks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	EngineId  = [4]byte{'a', 'u', 'r', 'a'}
	KeyTypeId = [4]byte{'a', 'u', 'r', 'a'}
)

var (
	errSlotMustIncrease      = errors.New("Slot must increase")
	errSlotDurationZero      = errors.New("Aura slot duration cannot be zero.")
	errTimestampSlotMismatch = errors.New("Timestamp slot must match `CurrentSlot`")
	errEmptyAuthorities      = errors.New("empty storage authorities")
	errZeroAuthorities       = errors.New("zero storage authorities")
)

type AuraModule interface {
	primitives.Module

	KeyType() primitives.PublicKeyType
	KeyTypeId() [4]byte
	OnTimestampSet(now sc.U64) error
	SlotDuration() sc.U64
	FindAuthor(digests sc.Sequence[primitives.DigestPreRuntime]) (sc.Option[sc.U32], error)
	StorageAuthorities() (sc.Sequence[primitives.Sr25519PublicKey], error)
	StorageAuthoritiesBytes() (sc.Option[sc.Sequence[sc.U8]], error)
	StorageCurrentSlot() (sc.U64, error)
}

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index              sc.U8
	config             *Config
	storage            *storage
	constants          *consts
	mdGenerator        *primitives.MetadataTypeGenerator
	logDepositor       system.LogDepositor
	disabledValidators primitives.DisabledValidators
	logger             log.Logger
}

func New(index sc.U8, config *Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.Logger) Module {
	storage := newStorage()

	return Module{
		index:              index,
		config:             config,
		storage:            storage,
		constants:          newConstants(config.DbWeight, config.MinimumPeriod),
		mdGenerator:        mdGenerator,
		logDepositor:       config.LogDepositor,
		disabledValidators: config.DisabledValidators,
		logger:             logger,
	}
}

func (m Module) GetIndex() sc.U8 {
	return m.index
}

func (m Module) name() sc.Str {
	return "Aura"
}

func (m Module) Functions() map[sc.U8]primitives.Call {
	return map[sc.U8]primitives.Call{}
}

func (m Module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m Module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

func (m Module) KeyType() primitives.PublicKeyType {
	return m.config.KeyType
}

func (m Module) KeyTypeId() [4]byte {
	return KeyTypeId
}

func (m Module) DecodeKey(buffer *bytes.Buffer) (primitives.Sr25519PublicKey, error) {
	key, err := primitives.DecodeSr25519PublicKey(buffer)
	if err != nil {
		return primitives.Sr25519PublicKey{}, err
	}

	return key, nil
}

func (m Module) OnInitialize(_ sc.U64) (primitives.Weight, error) {
	slot, err := m.currentSlotFromDigests()
	if err != nil {
		return primitives.Weight{}, err
	}

	if slot.HasValue {
		newSlot := slot.Value

		currentSlot, err := m.storage.CurrentSlot.Get()
		if err != nil {
			return primitives.Weight{}, err
		}

		if currentSlot >= newSlot {
			return primitives.Weight{}, errSlotMustIncrease
		}

		m.storage.CurrentSlot.Put(newSlot)

		totalAuthorities, err := m.storage.Authorities.DecodeLen()
		if err != nil {
			return primitives.Weight{}, err
		}
		if totalAuthorities.HasValue {
			authorityIndex := currentSlot % totalAuthorities.Value

			if m.disabledValidators != nil && m.disabledValidators.IsDisabled(sc.U32(authorityIndex)) {
				m.logger.Criticalf("Validator with index [%d] is disabled and should not be attempting to author blocks.", authorityIndex)
			}
		}

		return m.constants.DbWeight.ReadsWrites(2, 1), nil
	} else {
		return m.constants.DbWeight.Reads(1), nil
	}
}

func (m Module) OnTimestampSet(now sc.U64) error {
	slotDuration := m.SlotDuration()
	if slotDuration == 0 {
		return errSlotDurationZero
	}

	timestampSlot := now / slotDuration

	currentSlot, err := m.storage.CurrentSlot.Get()
	if err != nil {
		return err
	}
	if currentSlot != timestampSlot {
		return errTimestampSlotMismatch
	}
	return nil
}

func (m Module) OnGenesisSession(validators sc.Sequence[primitives.Validator]) error {
	authorities := sc.Sequence[primitives.Sr25519PublicKey]{}
	for _, validator := range validators {
		authorities = append(authorities, validator.AuthorityId)
	}

	if len(authorities) == 0 {
		return nil
	}

	return m.initializeAuthorities(authorities)
}

func (m Module) OnNewSession(isChanged bool, validators sc.Sequence[primitives.Validator], _ sc.Sequence[primitives.Validator]) error {
	if isChanged {
		nextAuthorities := sc.Sequence[primitives.Sr25519PublicKey]{}
		for _, validator := range validators {
			nextAuthorities = append(nextAuthorities, validator.AuthorityId)
		}
		lastAuthorities, err := m.StorageAuthorities()
		if err != nil {
			return err
		}

		if !reflect.DeepEqual(nextAuthorities, lastAuthorities) {
			if len(nextAuthorities) > int(m.config.MaxAuthorities) {
				m.logger.Warnf("next authorities list larger than maximum [%d], truncating", m.config.MaxAuthorities)
				return m.changeAuthorities(nextAuthorities[:m.config.MaxAuthorities])
			}

			return m.changeAuthorities(nextAuthorities)
		}
	}

	return nil
}

func (m Module) OnBeforeSessionEnding() {}

func (m Module) OnDisabled(validatorIndex sc.U32) {
	message := NewConsensusLogOnDisabled(validatorIndex).Bytes()
	log := primitives.NewDigestItemConsensusMessage(sc.BytesToFixedSequenceU8(KeyTypeId[:]), sc.BytesToSequenceU8(message))

	m.logDepositor.DepositLog(log)
}

// FindAuthor finds the author from the pre-runtime digests.
func (m Module) FindAuthor(digests sc.Sequence[primitives.DigestPreRuntime]) (sc.Option[sc.U32], error) {
	for _, preRuntime := range digests {
		if reflect.DeepEqual(sc.FixedSequenceU8ToBytes(preRuntime.ConsensusEngineId), EngineId[:]) {
			buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(preRuntime.Message))

			currentSlot, err := sc.DecodeU64(buffer)
			if err != nil {
				return sc.Option[sc.U32]{}, err
			}

			authoritiesLen, err := m.storageAuthoritiesLen()
			if err != nil {
				return sc.Option[sc.U32]{}, err
			}

			authorIndex := currentSlot % authoritiesLen

			return sc.NewOption[sc.U32](sc.U32(authorIndex)), nil
		}
	}

	return sc.NewOption[sc.U32](nil), nil
}

func (m Module) SlotDuration() sc.U64 {
	return m.constants.MinimumPeriod * 2
}

func (m Module) StorageAuthoritiesBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	return m.storage.Authorities.GetBytes()
}

func (m Module) StorageAuthorities() (sc.Sequence[primitives.Sr25519PublicKey], error) {
	return m.storage.Authorities.Get()
}

func (m Module) StorageCurrentSlot() (sc.U64, error) {
	return m.storage.CurrentSlot.Get()
}

// storageAuthoritiesLen fetches the length of the storage authorities.
// Returns err if value is empty or 0.
func (m Module) storageAuthoritiesLen() (sc.U64, error) {
	totalAuthorities, err := m.storage.Authorities.DecodeLen()
	if err != nil {
		return 0, err
	}

	if !totalAuthorities.HasValue {
		return 0, errEmptyAuthorities
	}

	if totalAuthorities.Value == 0 {
		return 0, errZeroAuthorities
	}

	return totalAuthorities.Value, nil
}

func (m Module) Metadata() primitives.MetadataModule {
	dataV14 := primitives.MetadataModuleV14{
		Name:      m.name(),
		Storage:   m.metadataStorage(),
		Call:      sc.NewOption[sc.Compact](nil),
		CallDef:   sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:     m.index,
	}
	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

func (m Module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithParams(
			metadata.TypesAuraStorageAuthorities,
			"BoundedVec<T::AuthorityId, T::MaxAuthorities>",
			sc.Sequence[sc.Str]{"bounded_collection", "bounded_vec", "BoundedVec"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesSequencePubKeys),
				}), sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.TypesAuthorityId, "T"),
				primitives.NewMetadataEmptyTypeParameter("S"),
			}),

		primitives.NewMetadataTypeWithPath(metadata.TypesAuthorityId,
			"sp_consensus_aura sr25519 app_sr25519 Public",
			sc.Sequence[sc.Str]{"sp_consensus_aura", "sr25519", "app_sr25519", "Public"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{primitives.NewMetadataTypeDefinitionField(metadata.TypesSr25519PubKey)})),

		primitives.NewMetadataTypeWithPath(metadata.TypesSr25519PubKey,
			"sp_core sr25519 Public",
			sc.Sequence[sc.Str]{"sp_core", "sr25519", "Public"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{primitives.NewMetadataTypeDefinitionField(metadata.TypesFixedSequence32U8)})),

		primitives.NewMetadataType(metadata.TypesSequencePubKeys,
			"[]PublicKey",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesAuthorityId))),

		primitives.NewMetadataTypeWithPath(metadata.TypesAuraSlot,
			"sp_consensus_slots Slot",
			sc.Sequence[sc.Str]{"sp_consensus_slots", "Slot"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU64),
				})),

		// type 924
		primitives.NewMetadataType(metadata.TypesTupleSequenceU8KeyTypeId, "(Seq<U8>, KeyTypeId)",
			primitives.NewMetadataTypeDefinitionTuple(sc.Sequence[sc.Compact]{sc.ToCompact(metadata.TypesSequenceU8), sc.ToCompact(metadata.TypesKeyTypeId)})),

		// type 923
		primitives.NewMetadataType(metadata.TypesSequenceTupleSequenceU8KeyTypeId, "[]byte TupleSequenceU8KeyTypeId", primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleSequenceU8KeyTypeId))),

		// type 922
		primitives.NewMetadataTypeWithParam(metadata.TypesOptionTupleSequenceU8KeyTypeId, "Option<TupleSequenceU8KeyTypeId>", sc.Sequence[sc.Str]{"Option"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant(
					"None",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{},
					0,
					""),
				primitives.NewMetadataDefinitionVariant(
					"Some",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceTupleSequenceU8KeyTypeId),
					},
					1,
					""),
			}),
			primitives.NewMetadataTypeParameter(metadata.TypesSequenceTupleSequenceU8KeyTypeId, "T")),
	}
}

func (m Module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
		Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
			primitives.NewMetadataModuleStorageEntry(
				"Authorities",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAuraStorageAuthorities)),
				"The current authority set."),
			primitives.NewMetadataModuleStorageEntry(
				"CurrentSlot",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAuraSlot)),
				"The current slot of this block.   This will be set in `on_initialize`."),
		},
	})
}

// initializeAuthorities initialises the authorities to the storage state.
// Returns an error if authorities already exist in the storage.
// Returns an error if new authorities are more than the maximum allowed.
func (m Module) initializeAuthorities(authorities sc.Sequence[primitives.Sr25519PublicKey]) error {
	totalAuthorities, err := m.storage.Authorities.DecodeLen()
	if err != nil {
		return err
	}

	if totalAuthorities.HasValue {
		return errAuthoritiesAlreadyInitialized
	}

	if len(authorities) > int(m.config.MaxAuthorities) {
		return errAuthoritiesExceedMaxAuthorities
	}

	m.storage.Authorities.Put(authorities)

	return nil
}

func (m Module) changeAuthorities(authorities sc.Sequence[primitives.Sr25519PublicKey]) error {
	if len(authorities) == 0 {
		m.logger.Warn("Ignoring empty authority change.")
		return nil
	}

	m.storage.Authorities.Put(authorities)

	message := NewConsensusLogAuthoritiesChange(authorities).Bytes()
	log := primitives.NewDigestItemConsensusMessage(sc.BytesToFixedSequenceU8(KeyTypeId[:]), sc.BytesToSequenceU8(message))

	m.logDepositor.DepositLog(log)
	return nil
}

func (m Module) currentSlotFromDigests() (sc.Option[slot], error) {
	digest, err := m.config.SystemDigest()
	if err != nil {
		return sc.Option[slot]{}, err
	}

	preRuntimeDigests, err := digest.PreRuntimes()
	if err != nil {
		return sc.Option[slot]{}, err
	}

	for _, preRuntime := range preRuntimeDigests {
		if reflect.DeepEqual(sc.FixedSequenceU8ToBytes(preRuntime.ConsensusEngineId), EngineId[:]) {
			buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(preRuntime.Message))

			currentSlot, err := sc.DecodeU64(buffer)
			if err != nil {
				return sc.Option[slot]{}, err
			}

			return sc.NewOption[slot](currentSlot), nil
		}
	}

	return sc.NewOption[slot](nil), nil
}
