package parachain_system

import (
	"bytes"
	"errors"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"reflect"
)

const (
	FunctionSetValidationData = iota
	functionSudoSendUpwardMessage
)

const (
	name = sc.Str("ParachainSystem")
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
	index       sc.U8
	constants   consts
	config      Config
	storage     *storage
	hashing     io.Hashing
	functions   map[sc.U8]primitives.Call
	mdGenerator *primitives.MetadataTypeGenerator
	logger      log.Logger
}

func New(index sc.U8, config Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.Logger) Module {
	constants := newConstants(config.DbWeight)
	functions := make(map[sc.U8]primitives.Call)

	module := Module{
		index:       index,
		constants:   constants,
		config:      config,
		storage:     newStorage(config.Storage),
		hashing:     io.NewHashing(),
		mdGenerator: mdGenerator,
		logger:      logger,
	}

	functions[FunctionSetValidationData] = newCallSetValidationData(index, FunctionSetValidationData, module)
	functions[functionSudoSendUpwardMessage] = newCallSendUpwardMessage(index, functionSudoSendUpwardMessage, module)

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

	// After this point, the `RelevantMessagingState` in storage reflects the
	// unincluded segment.
	err = m.adjustEgressBandwidthLimits()
	if err != nil {
		return err
	}

	_ = totalBandwidthOut
	_ = hostConfig
	_ = relayUpgradeGoAhead

	// TODO: Handle messages

	return nil
}

func (m Module) StorageNewValidationCodeBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	return m.storage.NewValidationCode.GetBytes()
}

// ScheduleCodeUpgrade contains logic for parachain upgrade functionality.
func (m Module) ScheduleCodeUpgrade(code sc.Sequence[sc.U8]) error {
	if !m.storage.ValidationData.Exists() {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.index,
			Err:     sc.U32(ErrorValidationDataNotAvailable),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	urs, err := m.storage.UpgradeRestrictionSignal.Get()
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	if urs.HasValue {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.index,
			Err:     sc.U32(ErrorProhibitedByPolkadot),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if m.storage.PendingValidationCode.Exists() {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.index,
			Err:     sc.U32(ErrorOverlappingUpgrades),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if !m.storage.HostConfiguration.Exists() {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.index,
			Err:     sc.U32(ErrorHostConfigurationNotAvailable),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	cfg, err := m.storage.HostConfiguration.Get()
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	if len(code) > int(cfg.MaxCodeSize) {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.index,
			Err:     sc.U32(ErrorTooBig),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	m.notifyPolkadotOfPendingUpgrade(code)
	m.storage.PendingValidationCode.Put(code)
	m.config.systemModule.DepositEvent(newEventValidationFunctionStored(m.index))

	return nil
}

// MaybeDropIncludedAncestors drops blocks from the unincluded segment with respect to the latest parachain head.
func (m Module) MaybeDropIncludedAncestors(storageProof parachain.RelayChainStateProof, capacity parachain.UnincludedSegmentCapacity) (primitives.Weight, error) {
	weightUsed := primitives.WeightZero()

	// if the unincluded segment length is non-zero, then the parachain head must be present.
	paraHead := storageProof.ReadIncludedParaHeadHash()

	unincludedSegmentLen := sc.U64(0)
	optUnincludedSegmentLen, err := m.storage.UnincludedSegment.DecodeLen()
	if err != nil {
		return primitives.Weight{}, err
	}
	if optUnincludedSegmentLen.HasValue {
		unincludedSegmentLen = optUnincludedSegmentLen.Value
	}

	weightUsed = weightUsed.Add(m.config.DbWeight.Reads(1))
	expectIncludedParent := capacity.IsExpectingIncludedParent()
	var includedHead sc.FixedSequence[sc.U8]
	if paraHead.HasValue {
		hash := paraHead.Value
		if expectIncludedParent {
			parentHash, err := m.config.systemModule.StorageParentHash()
			if err != nil {
				return primitives.Weight{}, err
			}
			if reflect.DeepEqual(hash, parentHash.FixedSequence) {
				return primitives.Weight{}, errors.New("expected parent to be included")
			}
		}
		includedHead = paraHead.Value
	} else {
		if expectIncludedParent {
			parentHash, err := m.config.systemModule.StorageParentHash()
			if err != nil {
				return primitives.Weight{}, err
			}
			includedHead = parentHash.FixedSequence
		} else {
			return primitives.Weight{}, errors.New("included head not present in relay storage proof")
		}
	}

	unIncludedSegment, err := m.storage.UnincludedSegment.Get()
	if err != nil {
		return primitives.Weight{}, err
	}

	var dropped sc.Sequence[parachain.Ancestor]
	if len(unIncludedSegment.Ancestors) > 0 {
		index := 0
		for i, ancestor := range unIncludedSegment.Ancestors {
			if !ancestor.ParaHeadHash.HasValue {
				return primitives.Weight{}, errors.New("para head hash is updated during block initialisation")
			}
			headHash := ancestor.ParaHeadHash.Value
			if reflect.DeepEqual(headHash.FixedSequence, includedHead) {
				index = i + 1
			}
		}

		newAncestors := unIncludedSegment.Ancestors[index:]
		m.storage.UnincludedSegment.Put(parachain.UnincludedSegment{Ancestors: newAncestors})
		dropped = unIncludedSegment.Ancestors[:index]
	}
	weightUsed = weightUsed.Add(m.config.DbWeight.ReadsWrites(1, 1))

	newLen := int(unincludedSegmentLen) - len(dropped)

	if len(dropped) > 0 {
		aggrBytes, err := m.storage.AggregatedUnincludedSegment.GetBytes()
		if err != nil {
			return primitives.Weight{}, err
		}
		if !aggrBytes.HasValue {
			return primitives.Weight{}, errors.New("dropped part of the segment wasn't empty")
		}

		aggr, err := parachain.DecodeSegmentTracker(bytes.NewBuffer(sc.SequenceU8ToBytes(aggrBytes.Value)))
		if err != nil {
			return primitives.Weight{}, err
		}

		for _, block := range dropped {
			err := aggr.Subtract(&block)
			if err != nil {
				return primitives.Weight{}, err
			}

		}
		m.storage.AggregatedUnincludedSegment.Put(aggr)
		weightUsed = weightUsed.Add(m.config.DbWeight.ReadsWrites(1, 1))
	}

	if newLen >= int(capacity.Get()) {
		return primitives.Weight{}, errors.New("no space left for the block in the unincluded segment")
	}

	return weightUsed, nil
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

	bytesHeadData, err := m.storage.CustomValidationHeadData.GetBytes()
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

func (m Module) enqueueInboundDownwardMessages(expectedDmqMqcHead primitives.H256, downwardMessages sc.Sequence[parachain.InboundDownwardMessage]) (primitives.Weight, error) {
	dmCount := len(downwardMessages)

	dmqHead, err := m.storage.LastDmqMqcHead.Get()
	if err != nil {
		return primitives.Weight{}, err
	}

	weightUsed := enqueueInboundDownwardMessagesWeight(sc.U64(dmCount), m.config.DbWeight)

	if dmCount != 0 {
		m.config.systemModule.DepositEvent(newEventDownwardMessagesReceived(m.index, sc.U32(dmCount)))

		for _, downwardMessage := range downwardMessages {
			err := dmqHead.ExtendDownward(downwardMessage, m.hashing)
			if err != nil {
				return primitives.Weight{}, err
			}
		}
		// TODO(message queue):
		/*
			let bounded = downward_messages
							.iter()
							// Note: we are not using `.defensive()` here since that prints the whole value to
							// console. In case that the message is too long, this clogs up the log quite badly.
							.filter_map(|m| match BoundedSlice::try_from(&m.msg[..]) {
								Ok(bounded) => Some(bounded),
								Err(_) => {
									defensive!("Inbound Downward message was too long; dropping");
									None
								},
							});
						T::DmpQueue::handle_messages(bounded);
		*/
		m.storage.LastDmqMqcHead.Put(dmqHead)
		m.config.systemModule.DepositEvent(newEventDownwardMessagesProcessed(m.index, weightUsed, dmqHead.RelayHash))
	}

	// After hashing each message in the message queue chain submitted by the collator, we
	// should arrive to the MQC head provided by the relay chain.
	//
	// A mismatch means that at least some of the submitted messages were altered, omitted or
	// added improperly.
	if reflect.DeepEqual(expectedDmqMqcHead, dmqHead.RelayHash) {
		return primitives.Weight{}, errors.New("mismatching expected mqc head")
	}

	return weightUsed, nil
}

// enqueueInboundHorizontalMessages processes all inbound horizontal messages relayed by the collator.
func (m Module) enqueueInboundHorizontalMessages(
	ingressChannels sc.Sequence[parachain.Channel],
	horizontalMessages parachain.HorizontalMessages,
	relayParentNumber parachain.RelayChainBlockNumber) (primitives.Weight, error) {
	// TODO: process all inbound horizontal messages
	m.storage.HrmpWatermark.Put(relayParentNumber)

	return primitives.Weight{}, nil
}

// adjustEgressBandwidthLimits adjusts the `RelevantMessagingState` according to the bandwidth limits in the
// unincluded segment.
func (m Module) adjustEgressBandwidthLimits() error {
	bytesUnincludedSegment, err := m.storage.AggregatedUnincludedSegment.GetBytes()
	if err != nil {
		return err
	}
	if !bytesUnincludedSegment.HasValue {
		return nil
	}

	_, err = parachain.DecodeSegmentTracker(bytes.NewBuffer(sc.SequenceU8ToBytes(bytesUnincludedSegment.Value)))
	if err != nil {
		return err
	}

	bytesMessagingState, err := m.storage.RelevantMessagingState.GetBytes()
	if err != nil {
		return err
	}
	if !bytesMessagingState.HasValue {
		return err
	}

	// TODO: Update RelevantMessagingState

	return nil
}

// sendUpwardMessage puts a message in the `PendingUpwardMessages` storage item.
// The message will be later sent in `on_finalize`.
// Checks host configuration to see if message is too big.
// Increases the delivery fee factor if the queue is sufficiently (see
// [`ump_constants::THRESHOLD_FACTOR`]) congested.
func (m Module) sendUpwardMessage(data sc.Sequence[sc.U8]) error {
	bytesHostConfiguration, err := m.storage.HostConfiguration.GetBytes()
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	if bytesHostConfiguration.HasValue {
		cfg, err := parachain.DecodeAbridgeHostConfiguration(bytes.NewBuffer(sc.SequenceU8ToBytes(bytesHostConfiguration.Value)))
		if err != nil {
			return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		}
		if len(data) > int(cfg.MaxUpwardMessageSize) {
			return primitives.NewDispatchErrorOther("message too big")
		}
		threshold := int(cfg.MaxUpwardQueueSize) / thresholdFactor
		m.storage.PendingUpwardMessages.AppendItem(data)

		upwardMessages, err := m.storage.PendingUpwardMessages.Get()
		if err != nil {
			return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		}

		totalSize := 0
		for _, um := range upwardMessages {
			totalSize += len(um)
		}
		if totalSize > threshold {
			messageSizeFactor := len(data) * messageSizeFeeBase
			err := m.increaseFeeFactor(messageSizeFactor)
			if err != nil {
				return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
			}
		}
	} else {
		m.storage.PendingUpwardMessages.AppendItem(data)
	}

	hash := m.hashing.Blake256(sc.SequenceU8ToBytes(data))
	h256, err := primitives.NewH256(sc.BytesToFixedSequenceU8(hash)...)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	m.config.systemModule.DepositEvent(newEventUpwardMessageSent(m.index, sc.NewOption[primitives.H256](h256)))

	return nil
}

func (m Module) increaseFeeFactor(messageSizeFactor int) error {
	deliveryFactor, err := m.storage.UpwardDeliveryFeeFactor.Get()
	if err != nil {
		return err
	}

	multiplier := sc.SaturatingAddU128(sc.NewU128(exponentialFeeBase), sc.NewU128(messageSizeFactor))
	newDeliveryFactor := deliveryFactor.Mul(multiplier)

	m.storage.UpwardDeliveryFeeFactor.Put(newDeliveryFactor)

	return nil
}

func (m Module) notifyPolkadotOfPendingUpgrade(code sc.Sequence[sc.U8]) {
	m.storage.NewValidationCode.Put(code)
	m.storage.DidSetValidationCode.Put(true)
}

func (m Module) Metadata() primitives.MetadataModule {
	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesParachainSystemCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesParachainSystemCalls, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<ParachainSystem, Runtime>"),
				},
				m.index,
				"Call.ParachainSystem"),
		),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesParachainSystemEvents)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesParachainSystemEvents, "cumulus_pallets_parachain_system::Event<Runtime>"),
				},
				m.index,
				"Events.ParachainSystem")),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesParachainSystemErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesParachainSystemErrors),
				},
				m.index,
				"Errors.ParachainSystem"),
		),
		Index: m.index,
	}

	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

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
				"UnincludedSegment",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceAncestor)),
				"Latest included block descendants the runtime accepted. In other words, these are "+
					"ancestors of the currently executing block which have not been included in the observed relay-chain state."+
					"The segment length is limited by the capacity returned from the [`ConsensusHook`] configured in the pallet."),
			primitives.NewMetadataModuleStorageEntry(
				"AggregatedUnincludedSegment",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSegmentTracker)),
				"Storage field that keeps track of bandwidth used by the unincluded segment along with the"+
					"latest HRMP watermark. Used for limiting the acceptance of new blocks with respect to relay chain constraints.",
			),
			primitives.NewMetadataModuleStorageEntry(
				"PendingValidationCode",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceU8)),
				"In case of a scheduled upgrade, this storage field contains the validation code to be applied."+
					"As soon as the relay chain gives us the go-ahead signal, we will overwrite the "+
					"[`:code`][sp_core::storage::well_known_keys::CODE] which will result the next block process "+
					"with the new validation code. This concludes the upgrade process."),
			primitives.NewMetadataModuleStorageEntry(
				"NewValidationCode",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceU8)),
				"Validation code that is set by the parachain and is to be communicated to collator and consequently relay-chain."+
					"This will be cleared in `on_initialize` of each new block if no other pallet already set the value."),
			primitives.NewMetadataModuleStorageEntry(
				"ValidationData",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesPersistedValidationData)),
				"The [`PersistedValidationData`] set for this block. "+
					"This value is expected to be set only once per block and it's never stored in the trie."),
			primitives.NewMetadataModuleStorageEntry(
				"DidSetValidationCode",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesBool)),
				"Were the validation data set to notify the relay chain?"),
			primitives.NewMetadataModuleStorageEntry(
				"LastRelayChainBlockNumber",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU32)),
				"The relay chain block number associated with the last parachain block. This is updated in `on_finalize`."),
			primitives.NewMetadataModuleStorageEntry(
				"UpgradeRestrictionSignal",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesOptionUpgradeRestriction)),
				"An option which indicates if the relay-chain restricts signalling a validation code upgrade."+
					" In other words, if this is `Some` and [`NewValidationCode`] is `Some` then the produced "+
					"candidate will be invalid. "+
					"This storage item is a mirror of the corresponding value for the current parachain from the "+
					"elay-chain. This value is ephemeral which means it doesn't hit the storage. This value is"+
					" set after the inherent."),
			primitives.NewMetadataModuleStorageEntry(
				"UpgradeGoAhead",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesOptionUpgradeGoAhead)),
				"Optional upgrade go-ahead signal from the relay-chain."+
					" This storage item is a mirror of the corresponding value for the current parachain from the"+
					" relay-chain. This value is ephemeral which means it doesn't hit the storage. This value is"+
					" set after the inherent."),
			primitives.NewMetadataModuleStorageEntry(
				"RelayStateProof",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceSequenceU8)),
				"The state proof for the last relay parent block."+
					" This field is meant to be updated each block with the validation data inherent. Therefore,"+
					" before processing of the inherent, e.g. in `on_initialize` this data may be stale."+
					" This data is also absent from the genesis."),
			primitives.NewMetadataModuleStorageEntry(
				"RelevantMessagingState",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesMessagingStateSnapshot)),
				"The snapshot of some state related to messaging relevant to the current parachain as per"+
					" the relay parent. "+
					"This field is meant to be updated each block with the validation data inherent. Therefore,"+
					" before processing of the inherent, e.g. in `on_initialize` this data may be stale."+
					" This data is also absent from the genesis.",
			),
			primitives.NewMetadataModuleStorageEntry(
				"HostConfiguration",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesMessagingStateSnapshot)),
				"The parachain host configuration that was obtained from the relay parent."+
					"This field is meant to be updated each block with the validation data inherent. Therefore,"+
					" before processing of the inherent, e.g. in `on_initialize` this data may be stale."+
					" This data is also absent from the genesis.",
			),
			primitives.NewMetadataModuleStorageEntry(
				"LastDmqMqcHead",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesH256)),
				"The parachain host configuration that was obtained from the relay parent."+
					"The last downward message queue chain head we have observed."+
					" This value is loaded before and saved after processing inbound downward messages carried"+
					" by the system inherent.",
			),
			primitives.NewMetadataModuleStorageEntry(
				"ProcessedDownwardMessages",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU32)),
				"Number of downward messages processed in a block. This will be cleared in `on_initialize` of each new block."),
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
				"PendingUpwardMessages",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceSequenceU8)),
				"Upward messages that are still pending and not yet send to the relay chain."),
			primitives.NewMetadataModuleStorageEntry(
				"UpwardDeliveryFeeFactor",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU128)),
				"The factor to multiply the base delivery fee by for UMP."),
			primitives.NewMetadataModuleStorageEntry(
				"AnnouncedHrmpMessagesPerCandidate",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU32)),
				"The number of HRMP messages we observed in `on_initialize` and thus used that number for"+
					" announcing the weight of `on_initialize` and `on_finalize`."),
			primitives.NewMetadataModuleStorageEntry(
				"CustomValidationHeadData",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSequenceU8)),
				"A custom head data that should be returned as result of `validate_block`."),
		},
	})
}
