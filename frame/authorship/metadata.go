package authorship

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func (m module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
		Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
			primitives.NewMetadataModuleStorageEntry(
				"Author",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAddress32)),
				"Author of current block.",
			),
		},
	})
}

func (m module) Metadata() primitives.MetadataModule {
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
