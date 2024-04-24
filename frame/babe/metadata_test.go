package babe

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func Test_Metadata(t *testing.T) {
	target := setupModule()

	result := target.Metadata()

	moduleV14 := primitives.MetadataModuleV14{
		Name: "Babe",
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: "Babe",
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"Authorities",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBoundedVecAuthority)),
					"Current epoch authorities.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"AuthorVrfRandomness",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesOptionFixedSequence32U8)),
					"This field should always be populated during block processing unless secondary plain slots are enabled (which don't contain a VRF output). It is set in `on_finalize`, before it will contain the value from the last block.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"CurrentSlot",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSlot)),
					"Current slot number.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"EpochConfig",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBabeEpochConfiguration)),
					"The configuration for the current epoch. Should never be `None` as it is initialized in genesis.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"EpochIndex",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU64)),
					"Current epoch index.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"EpochStart",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesTuple2U64)),
					"The block numbers when the last and current epoch have started, respectively `N-1` and `N`. NOTE: We track this is in order to annotate the block number when a given pool of entropy was fixed (i.e. it was known to chain observers). Since epochs are defined in slots, which may be skipped, the block numbers may not line up with the slot numbers.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"GenesisSlot",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSlot)),
					"The slot at which the first epoch actually started. This is 0 until the first block of the chain.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"Initialized",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesOptionPreDigest)),
					"Temporary value (cleared at block finalization) which is `Some` if per-block initialization has already been called for current block.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"Lateness",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU64)),
					"How late the current block is compared to its parent. This entry is populated as part of block execution and is cleaned up on block finalization. Querying this storage entry outside of block execution context should always yield zero.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"NextAuthorities",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBoundedVecAuthority)),
					"Authorities set scheduled to be used with the next session",
				),
				primitives.NewMetadataModuleStorageEntry(
					"NextEpochConfig",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBabeEpochConfiguration)),
					"The configuration for the next epoch, `None` if the config will not change (you can fallback to `EpochConfig` instead in that case).",
				),
				primitives.NewMetadataModuleStorageEntry(
					"NextRandomness",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesFixedSequence32U8)),
					"Next epoch randomness.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"PendingEpochConfigChange",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBabeNextConfigDescriptor)),
					"Pending epoch configuration change that will be applied when the next epoch is enacted.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"Randomness",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesFixedSequence32U8)),
					"The epoch randomness for the *current* epoch.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"SegmentIndex",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU32)),
					"Randomness under construction.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"SkippedEpochs",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBabeBoundedVecSkippedEpoch)),
					"A list of the last 100 skipped epochs and the corresponding session index when the epoch was skipped. This is only used for validating equivocation proofs. An equivocation proof must contains a key-ownership proof for a given session, therefore we need a way to tie together sessions and epoch indices, i.e. we need to validate that a validator was the owner of a given key on a given session, and what the active epoch index was during that session.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"UnderConstruction",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
						sc.ToCompact(metadata.PrimitiveTypesU32),
						sc.ToCompact(metadata.TypesFixedSequence32U8),
					),
					"TWOX-NOTE: `SegmentIndex` is an increasing integer, so this is okay.",
				),
			},
		}),
		Call: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesBabeCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				"Babe",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(
						metadata.TypesBabeCalls,
						"self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Babe, Runtime>",
					),
				},
				moduleId,
				"Call.Babe",
			),
		),
		Event:    sc.NewOption[sc.Compact](nil),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{
			primitives.NewMetadataModuleConstant(
				"EpochDuration",
				sc.ToCompact(metadata.PrimitiveTypesU64),
				sc.BytesToSequenceU8(epochDuration.Bytes()),
				"The amount of time, in slots, that each epoch should last. NOTE: Currently it is not possible to change the epoch duration after the chain has started. Attempting to do so will brick block production.",
			),
			primitives.NewMetadataModuleConstant(
				"ExpectedBlockTime",
				sc.ToCompact(metadata.PrimitiveTypesU64),
				sc.BytesToSequenceU8(timestampMinimumPeriod.Bytes()),
				"The expected average block time at which BABE should be creating blocks. Since BABE is probabilistic it is not trivial to figure out what the expected average block time should be based on the slot duration and the security parameter `c` (where `1 - c` represents the probability of a slot being empty).",
			),
			primitives.NewMetadataModuleConstant(
				"MaxAuthorities",
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.BytesToSequenceU8(maxAuthorities.Bytes()),
				"Max number of authorities allowed",
			),
		},
		Error: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesBabeErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				"Babe",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesBabeErrors),
				},
				moduleId,
				"Errors.Babe",
			),
		),
		Index: moduleId,
	}

	expectMetadataModule := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: moduleV14,
	}

	assert.Equal(t, expectMetadataModule, result)
}
