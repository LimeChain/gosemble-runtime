package parachain_system

import (
	"bytes"
	"errors"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	FunctionSetValidationData = iota
	name                      = sc.Str("ParachainSystem")
)

var (
	inherentIdentifier = [8]byte{'s', 'y', 's', 'i', '1', '3', '3', '7'}
)

var (
	errInherentNotProvided             = errors.New("Parachain system inherent data must be provided.")
	errInherentDataNotCorrectlyEncoded = errors.New("Parachain system inherent data not correctly encoded.")
)

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index     sc.U8
	constants consts
	config    Config
	storage   *storage
	functions map[sc.U8]primitives.Call
	logger    log.Logger
}

func New(index sc.U8, config Config, logger log.Logger) Module {
	constants := newConstants(config.DbWeight)
	functions := make(map[sc.U8]primitives.Call)

	module := Module{
		index:     index,
		constants: constants,
		config:    config,
		storage:   newStorage(config.Storage),
		logger:    logger,
	}

	functions[FunctionSetValidationData] = newCallSetValidationData(index, FunctionSetValidationData, module)

	module.functions = functions

	return module
}

func (m Module) GetIndex() sc.U8 {
	return m.index
}

func (m Module) name() sc.Str {
	return name
}

func (m Module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m Module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m Module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.DefaultValidTransaction(), nil
}

func (m Module) CreateInherent(inherent primitives.InherentData) (sc.Option[primitives.Call], error) {
	inherentData := inherent.Get(inherentIdentifier)

	if inherentData == nil {
		return sc.Option[primitives.Call]{}, errInherentNotProvided
	}

	buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(inherentData))
	data, err := DecodeParachainInherentData(buffer)
	if err != nil {
		return sc.Option[primitives.Call]{}, errInherentDataNotCorrectlyEncoded
	}

	data, err = m.DropProcessedMessagesFromInherent(data)
	if err != nil {
		return sc.Option[primitives.Call]{}, err
	}

	function := newCallSetValidationDataWithArgs(m.index, FunctionSetValidationData, sc.NewVaryingData(data))

	return sc.NewOption[primitives.Call](function), nil
}

func (m Module) CheckInherent(call primitives.Call, inherent primitives.InherentData) error {
	if !m.IsInherent(call) {
		return NewInherentErrorInvalid()
	}

	return nil
}

func (m Module) InherentIdentifier() [8]byte { return inherentIdentifier }

func (m Module) IsInherent(call primitives.Call) bool {
	return call.ModuleIndex() == m.index && call.FunctionIndex() == FunctionSetValidationData
}

func (m Module) OnInitialize(_ sc.U64) (primitives.Weight, error) {
	weight := primitives.WeightZero()

	didSetValidationCode, err := m.storage.DidSetValidationCode.Get()
	if err != nil {
		return primitives.WeightZero(), err
	}
	if !didSetValidationCode {
		m.storage.NewValidationCode.Clear()
		weight = weight.Add(m.config.DbWeight.Writes(1))
	}

	// The parent hash was unknown during block finalization. Update it here.
	unincludedSegment, err := m.storage.UnincludedSegment.Get()
	if err != nil {
		return primitives.WeightZero(), err
	}
	if len(unincludedSegment.Ancestors) > 0 {
		// Ancestor is the latest finalized block, thus current parent is
		// its output head.
		last := unincludedSegment.Ancestors[len(unincludedSegment.Ancestors)-1]
		parentHash, err := m.config.systemModule.StorageParentHash()
		if err != nil {
			return primitives.WeightZero(), err
		}
		last.ParaHeadHash = sc.NewOption[primitives.H256](primitives.H256{FixedSequence: parentHash.FixedSequence})
		unincludedSegment.Ancestors[len(unincludedSegment.Ancestors)-1] = last

		m.storage.UnincludedSegment.Put(unincludedSegment)

		weight = weight.Add(m.config.DbWeight.ReadsWrites(1, 1))
		// Weight used during finalization.
		weight = weight.Add(m.config.DbWeight.ReadsWrites(3, 2))
	}

	// Remove the validation from the old block.
	m.storage.ValidationData.Clear()
	m.storage.ProcessedDownwardMessages.Clear()
	m.storage.HrmpWatermark.Clear()
	m.storage.UpwardMessages.Clear()
	m.storage.HrmpOutboundMessages.Clear()
	m.storage.CustomValidationHeadData.Clear()

	weight = weight.Add(m.constants.DbWeight.Writes(6))

	weight = weight.Add(m.config.DbWeight.ReadsWrites(1, 1))
	hostConfig, err := m.storage.HostConfiguration.Get()
	if err != nil {
		return primitives.WeightZero(), err
	}
	m.storage.AnnouncedHrmpMessagesPerCandidate.Put(hostConfig.MaxHrmpMessageNumPerCandidate)

	// NOTE that the actual weight consumed by `on_finalize` may turn out lower.
	weight = weight.Add(m.config.DbWeight.ReadsWrites(3, 4))

	// Weight for updating the last relay chain block number in `on_finalize`.
	weight = weight.Add(m.config.DbWeight.ReadsWrites(1, 1))

	// Weight for adjusting the unincluded segment in `on_finalize`.
	weight = weight.Add(m.config.DbWeight.ReadsWrites(6, 3))

	// Always try to read `UpgradeGoAhead` in `on_finalize`.
	weight = weight.Add(m.config.DbWeight.Reads(1))

	return weight, nil
}

func (m Module) OnFinalize(_ sc.U64) error {
	m.storage.DidSetValidationCode.Clear()
	m.storage.UpgradeRestrictionSignal.Clear()

	relayUpgradeGoAhead, err := m.storage.UpgradeGoAhead.Take()
	if err != nil {
		return err
	}

	validationData, err := m.storage.ValidationData.Get()
	if err != nil {
		return err
	}

	m.storage.LastRelayChainBlockNumber.Put(validationData.RelayParentNumber)

	optionHostConfig, err := m.storage.HostConfiguration.GetBytes()
	if err != nil {
		return err
	}
	if !optionHostConfig.HasValue {
		return errors.New("host configuration is promised to be set until `onFinalize`")
	}

	hostConfig, err := parachain.DecodeAbridgeHostConfiguration(bytes.NewBuffer(sc.SequenceU8ToBytes(optionHostConfig.Value)))
	if err != nil {
		return err
	}

	optionMessagingState, err := m.storage.RelevantMessagingState.GetBytes()
	if err != nil {
		return err
	}
	if !optionMessagingState.HasValue {
		return errors.New("relevant messaging state is promised to be set until `onFinalize`")
	}

	messagingState, err := parachain.DecodeMessagingStateSnapshot(bytes.NewBuffer(sc.SequenceU8ToBytes(optionMessagingState.Value)))
	if err != nil {
		return err
	}

	totalBandwidthOut := parachain.NewOutboundBandwidthLimitsFromMessagingStateSnapshot(messagingState)

	m.adjustEgressBandwidthLimits()

	_ = totalBandwidthOut
	_ = hostConfig
	_ = relayUpgradeGoAhead

	return nil
}

func (m Module) adjustEgressBandwidthLimits() {
	// TODO:
}

func (m Module) StorageNewValidationCodeBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	return m.storage.NewValidationCode.GetBytes()
}

func (m Module) MaybeDropIncludedAncestors(storageProof parachain.RelayChainStateProof, capacity parachain.UnincludedSegmentCapacity) primitives.Weight {
	// TODO:
	return primitives.WeightZero()
}

func (m Module) CollectCollationInfo(header primitives.Header) (CollationInfo, error) {
	hrmpWatermark, err := m.storage.HrmpWatermark.Get()
	if err != nil {
		return CollationInfo{}, err
	}

	horizontalMessages, err := m.storage.HrmpOutboundMessages.Get()
	if err != nil {
		return CollationInfo{}, err
	}

	upwardMessages, err := m.storage.UpwardMessages.Get()
	if err != nil {
		return CollationInfo{}, err
	}

	processedDownwardMessages, err := m.storage.ProcessedDownwardMessages.Get()
	if err != nil {
		return CollationInfo{}, err
	}

	newValidationCode, err := m.storage.NewValidationCode.GetBytes()
	if err != nil {
		return CollationInfo{}, err
	}

	bytesHeadData, err := m.storage.CustomValidationHeadData.Get()
	if err != nil {
		return CollationInfo{}, err
	}
	var headData sc.Sequence[sc.U8]
	if bytesHeadData.HasValue {
		headData = bytesHeadData.Value
	} else {
		headData = sc.BytesToSequenceU8(header.Bytes())
	}

	return CollationInfo{
		UpwardMessages:            upwardMessages,
		HorizontalMessages:        horizontalMessages,
		ValidationCode:            newValidationCode,
		ProcessedDownwardMessages: processedDownwardMessages,
		HrmpWatermark:             hrmpWatermark,
		HeadData:                  headData,
	}, nil
}

func (m Module) DropProcessedMessagesFromInherent(parachainInherent ParachainInherentData) (ParachainInherentData, error) {
	relayChainBlockNumber, err := m.storage.LastRelayChainBlockNumber.Get()
	if err != nil {
		return ParachainInherentData{}, err
	}

	downwardMessages := sc.Sequence[parachain.InboundDownwardMessage]{}
	for _, downWardMessage := range parachainInherent.DownwardMessages {
		if downWardMessage.SentAt > relayChainBlockNumber {
			downwardMessages = append(downwardMessages, downWardMessage)
		}
	}

	horizontalMessages := parachainInherent.HorizontalMessages.UnprocessedMessages(relayChainBlockNumber)

	return ParachainInherentData{
		ValidationData:     parachainInherent.ValidationData,
		RelayChainState:    parachainInherent.RelayChainState,
		DownwardMessages:   downwardMessages,
		HorizontalMessages: horizontalMessages,
	}, nil
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

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

func (m Module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
		Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
			primitives.NewMetadataModuleStorageEntry(
				"NewValidationCode",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceU8)),
				"Validation code that is set by the parachain and is to be communicated to collator and consequently relay-chain."+
					"This will be cleared in `on_initialize` of each new block if no other pallet already set the value."),
			primitives.NewMetadataModuleStorageEntry(
				"ProcessedDownwardMessages",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU32)),
				"Number of downward messages processed in a block. This will eb cleared in `on_initialize` of each new block."),
			primitives.NewMetadataModuleStorageEntry(
				"HrmpWatermark",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU32)),
				"HRMP watermark that was set in a block. This will be cleared in `on_initialize` of each new block."),
			primitives.NewMetadataModuleStorageEntry(
				"HrmpOutboundMessages",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesParachainOutboundHrmpMessages)),
				"HRMP messages that were sent in a block. This will be cleared in `on_initialize` of each block"),
			primitives.NewMetadataModuleStorageEntry(
				"UpwardMessages",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceSequenceU8)),
				"Upward messages that were sent in a block. This will be cleared in `on_initialize` of each new block."),
			primitives.NewMetadataModuleStorageEntry(
				"CustomValidationHeadData",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesOptionSequenceU8)),
				"A custom head data that should be returned as result of `validate_block`."),
		},
	})
}
