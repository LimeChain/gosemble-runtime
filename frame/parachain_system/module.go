package parachain_system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/hooks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index     sc.U8
	constants consts
	storage   *storage
}

func New(index sc.U8, config Config) Module {
	constants := newConstants(config.DbWeight)
	return Module{
		index:     index,
		constants: constants,
		storage:   newStorage(),
	}
}

func (m Module) GetIndex() sc.U8 {
	return m.index
}

func (m Module) name() sc.Str {
	return "ParachainSystem"
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

func (m Module) OnInitialize(_ sc.U64) (primitives.Weight, error) {
	weight := primitives.WeightZero()

	// TODO: add validation data
	m.storage.ProcessedDownwardMessages.Clear()
	m.storage.HrmpWatermark.Clear()
	m.storage.UpwardMessages.Clear()
	m.storage.HrmpOutboundMessages.Clear()
	m.storage.CustomValidationHeadData.Clear()

	weight = weight.Add(m.constants.DbWeight.Writes(6))

	return weight, nil
}

func (m Module) StorageNewValidationCodeBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	return m.storage.NewValidationCode.GetBytes()
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
