package aura_ext

import (
	"bytes"
	"fmt"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/aura"
	"github.com/LimeChain/gosemble/frame/executive"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index      sc.U8
	config     Config
	constants  consts
	storage    *storage
	auraModule aura.AuraModule
	executive  executive.Module
	logger     log.RuntimeLogger
}

func New(index sc.U8, config Config, aura aura.AuraModule, logger log.RuntimeLogger) Module {
	storage := newStorage(config.Storage)
	constants := newConstants(config.DbWeight)

	return Module{
		index:      index,
		config:     config,
		constants:  constants,
		storage:    storage,
		auraModule: aura,
		logger:     logger,
	}
}

func (m Module) GetIndex() sc.U8 {
	return m.index
}

func (m Module) name() sc.Str {
	return "AuraExt"
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
	// Fetch the authorities once to get them into the storage proof of the PoV.
	_, err := m.storage.Authorities.Get()
	if err != nil {
		return primitives.Weight{}, err
	}

	newSlot, err := m.auraModule.StorageCurrentSlot()
	if err != nil {
		return primitives.Weight{}, err
	}
	m.logger.Trace("pass OnInitialize module")

	bytesSlotInfo, err := m.storage.SlotInfo.GetBytes()
	if err != nil {
		return primitives.Weight{}, err
	}

	if !bytesSlotInfo.HasValue {
		m.logger.Tracef("Slot info has no previously set value, setting values [%d], [%d]", newSlot, 1)
		m.storage.SlotInfo.Put(SlotInfo{
			Slot:     newSlot,
			Authored: 1,
		})
	} else {
		slot, err := DecodeSlotInfo(bytes.NewBuffer(sc.SequenceU8ToBytes(bytesSlotInfo.Value)))
		if err != nil {
			return primitives.Weight{}, err
		}

		if slot.Slot == newSlot {
			m.storage.SlotInfo.Put(SlotInfo{
				Slot:     slot.Slot,
				Authored: slot.Authored + 1,
			})
		} else if slot.Slot < newSlot {
			m.storage.SlotInfo.Put(SlotInfo{
				Slot:     newSlot,
				Authored: 1,
			})
		} else {
			return primitives.Weight{}, fmt.Errorf("slot moved backwards current [%d], new [%d]", slot.Slot, newSlot)
		}
	}

	return m.constants.DbWeight.ReadsWrites(2, 1), nil
}

func (m Module) OnFinalize(_ sc.U64) error {
	authorities, err := m.auraModule.StorageAuthorities()
	if err != nil {
		return err
	}

	m.storage.Authorities.Put(authorities)

	return nil
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
