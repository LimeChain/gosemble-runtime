package babe

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	optionNoneIdx sc.U8 = iota
	optionSomeIdx
)

func (m module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{

		// 161
		primitives.NewMetadataType(
			metadata.TypesRationalValueU64,
			"RationalValue",
			primitives.NewMetadataTypeDefinitionTuple(
				sc.Sequence[sc.Compact]{
					sc.ToCompact(metadata.PrimitiveTypesU64),
					sc.ToCompact(metadata.PrimitiveTypesU64),
				},
			),
		),

		// 158
		primitives.NewMetadataTypeWithPath(
			metadata.TypesSlot,
			"Slot",
			sc.Sequence[sc.Str]{"sp_consensus_slots", "Slot"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU64),
				},
			),
		),

		// 157
		primitives.NewMetadataTypeWithPath(
			metadata.TypesSr25519PubKey,
			"sr25519::Public",
			sc.Sequence[sc.Str]{"sp_consensus_babe", "app", "Public"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesFixedSequence32U8),
				},
			),
		),

		// 521
		primitives.NewMetadataType(
			metadata.TypesAuthority,
			"Authority",
			primitives.NewMetadataTypeDefinitionTuple(
				sc.Sequence[sc.Compact]{
					sc.ToCompact(metadata.TypesSr25519PubKey),
					sc.ToCompact(metadata.PrimitiveTypesU64),
				},
			),
		),

		// 522
		primitives.NewMetadataType(
			metadata.TypesSequenceAuthority,
			"SequenceAuthority",
			primitives.NewMetadataTypeDefinitionSequence(
				sc.ToCompact(metadata.TypesAuthority),
			),
		),

		// 520
		primitives.NewMetadataTypeWithParams(
			metadata.TypesBoundedVecAuthority,
			"WeakBoundedVec<(AuthorityId, BabeAuthorityWeight), T::MaxAuthorities>",
			sc.Sequence[sc.Str]{"bounded_collections", "weak_bounded_vec", "WeakBoundedVec"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSequenceAuthority, "Vec<T>"),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.TypesAuthority, "T"),
				primitives.NewMetadataEmptyTypeParameter("S"),
			},
		),

		// 95
		primitives.NewMetadataTypeWithParam(
			metadata.TypesOptionFixedSequence32U8,
			"Option<[u8; 32]>",
			sc.Sequence[sc.Str]{"Option"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"None",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						optionNoneIdx,
						"",
					),
					primitives.NewMetadataDefinitionVariant(
						"Some",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesFixedSequence32U8),
						},
						optionSomeIdx,
						"",
					),
				},
			),
			primitives.NewMetadataTypeParameter(metadata.TypesFixedSequence32U8, "T"),
		),

		// 533
		primitives.NewMetadataType(
			metadata.TypesBabeSkippedEpoch,
			"SkippedEpochs",
			primitives.NewMetadataTypeDefinitionTuple(
				sc.Sequence[sc.Compact]{
					sc.ToCompact(metadata.PrimitiveTypesU64),
					sc.ToCompact(metadata.PrimitiveTypesU32),
				},
			),
		),

		// 534
		primitives.NewMetadataType(
			metadata.TypesBabeFixedSequenceSkippedEpoch,
			"Seq<SkippedEpoch>",
			primitives.NewMetadataTypeDefinitionSequence(
				sc.ToCompact(metadata.TypesBabeSkippedEpoch),
			),
		),

		// 532
		primitives.NewMetadataTypeWithParams(
			metadata.TypesBabeBoundedVecSkippedEpoch,
			"BoundedVec<SkippedEpoch, ConstU32>",
			sc.Sequence[sc.Str]{"bounded_collections", "bounded_vec", "BoundedVec"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesBabeFixedSequenceSkippedEpoch, "Vec<T>"),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.TypesBabeSkippedEpoch, "T"),
				primitives.NewMetadataEmptyTypeParameter("S"),
			},
		),

		// 531
		primitives.NewMetadataTypeWithPath(
			metadata.TypesBabeEpochConfiguration,
			"EpochConfiguration",
			sc.Sequence[sc.Str]{"sp_consensus_babe", "BabeEpochConfiguration"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesRationalValueU64, "(u64, u64)", "c"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAllowedSlots, "AllowedSlots", "allowed_slots"),
				},
			),
		),

		// 160
		primitives.NewMetadataTypeWithPath(
			metadata.TypesBabeNextConfigDescriptor,
			"NextConfigDescriptor",
			sc.Sequence[sc.Str]{"sp_consensus_babe", "digests", "NextConfigDescriptor"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"V1",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesRationalValueU64, "(u64, u64)", "c"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAllowedSlots, "AllowedSlots", "allowed_slots"),
						},
						NextConfigDescriptorV1,
						"",
					),
				},
			),
		),

		// 528
		primitives.NewMetadataTypeWithPath(
			metadata.TypesVrfSignature,
			"VrfSignature",
			sc.Sequence[sc.Str]{"sp_core", "sr25519	", "vrf", "VrfSignature"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesFixedSequence32U8, "VrfPreOutput", "pre_output"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesFixedSequence64U8, "VrfProof", "proof"),
				},
			),
		),

		// 527
		primitives.NewMetadataTypeWithPath(
			metadata.TypesBabePrimaryPreDigest,
			"Primary",
			sc.Sequence[sc.Str]{"sp_consensus_babe", "digests", "PrimaryPreDigest"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "super::AuthorityIndex", "authority_index"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSlot, "Slot", "slot"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesVrfSignature, "VrfSignature", "vrf_signature"),
				},
			),
		),

		// 529
		primitives.NewMetadataTypeWithPath(
			metadata.TypesBabeSecondaryPlainPreDigest,
			"SecondaryPlain",
			sc.Sequence[sc.Str]{"sp_consensus_babe", "digests", "SecondaryPlainPreDigest"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "super::AuthorityIndex", "authority_index"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSlot, "Slot", "slot"),
				},
			),
		),

		// 530
		primitives.NewMetadataTypeWithPath(
			metadata.TypesBabeSecondaryVRFPreDigest,
			"SecondaryVRF",
			sc.Sequence[sc.Str]{"sp_consensus_babe", "digests", "SecondaryVRFPreDigest"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "super::AuthorityIndex", "authority_index"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSlot, "Slot", "slot"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesVrfSignature, "VrfSignature", "vrf_signature"),
				},
			),
		),

		// 526
		primitives.NewMetadataTypeWithPath(
			metadata.TypesPreDigest,
			"PreDigest",
			sc.Sequence[sc.Str]{"sp_consensus_babe", "digests", "PreDigest"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Primary",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesBabePrimaryPreDigest, "PrimaryPreDigest"),
						},
						Primary,
						"",
					),
					primitives.NewMetadataDefinitionVariant(
						"SecondaryPlain",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesBabeSecondaryPlainPreDigest, "SecondaryPlainPreDigest"),
						},
						SecondaryPlain,
						"",
					),
					primitives.NewMetadataDefinitionVariant(
						"SecondaryVRF",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesBabeSecondaryVRFPreDigest, "SecondaryVRFPreDigest"),
						},
						SecondaryVRF,
						"",
					),
				},
			),
		),

		// 525
		primitives.NewMetadataTypeWithParam(
			metadata.TypesOptionPreDigest,
			"Option<PreDigest>",
			sc.Sequence[sc.Str]{"Option"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant("None", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, optionNoneIdx, ""),
					primitives.NewMetadataDefinitionVariant("Some", sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionField(metadata.TypesPreDigest)}, optionSomeIdx, ""),
				},
			),
			primitives.NewMetadataTypeParameter(metadata.TypesPreDigest, "T"),
		),

		// 162
		primitives.NewMetadataTypeWithPath(
			metadata.TypesAllowedSlots,
			"AllowedSlots",
			sc.Sequence[sc.Str]{"sp_consensus_babe", "AllowedSlots"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant("PrimarySlots", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, babetypes.PrimarySlots, "BabePrimarySlots"),
					primitives.NewMetadataDefinitionVariant("PrimaryAndSecondaryPlainSlots", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, babetypes.PrimaryAndSecondaryPlainSlots, "PrimaryAndSecondaryPlainSlots"),
					primitives.NewMetadataDefinitionVariant("PrimaryAndSecondaryVRFSlots", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, babetypes.PrimaryAndSecondaryVRFSlots, "PrimaryAndSecondaryVRFSlots"),
				},
			),
		),

		// 535
		primitives.NewMetadataTypeWithParams(
			metadata.TypesBabeErrors,
			"Babe Errors",
			sc.Sequence[sc.Str]{"pallet_babe", "pallet", "Error"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"InvalidEquivocationProof",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorInvalidEquivocationProof,
						"An equivocation proof provided as part of an equivocation report is invalid.",
					),
					primitives.NewMetadataDefinitionVariant(
						"InvalidKeyOwnershipProof",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorInvalidKeyOwnershipProof,
						"A key ownership proof provided as part of an equivocation report is invalid.",
					),
					primitives.NewMetadataDefinitionVariant(
						"DuplicateOffenceReport",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorDuplicateOffenceReport,
						"A given equivocation report is valid but already previously reported.",
					),
					primitives.NewMetadataDefinitionVariant(
						"InvalidConfiguration",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorInvalidConfiguration,
						"Submitted configuration is invalid.",
					),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
			},
		),

		// 153
		primitives.NewMetadataTypeWithParam(
			metadata.TypesBabeCalls,
			"Babe calls",
			sc.Sequence[sc.Str]{"pallet_babe", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"plan_config_change",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesBabeNextConfigDescriptor, "config"),
						},
						functionPlanConfigChangeIndex,
						"Plan an epoch config change. The epoch config change is recorded and will be enacted on the next call to `enact_epoch_change`. The config will be activated one epoch after. Multiple calls to this method will replace any existing planned config change that had not been enacted yet.",
					),
				},
			),
			primitives.NewMetadataEmptyTypeParameter("T"),
		),
	}
}

func (m module) metadataConstants() sc.Sequence[primitives.MetadataModuleConstant] {
	return sc.Sequence[primitives.MetadataModuleConstant]{
		primitives.NewMetadataModuleConstant(
			"EpochDuration",
			sc.ToCompact(metadata.PrimitiveTypesU64),
			sc.BytesToSequenceU8(m.constants.EpochDuration.Bytes()),
			"The amount of time, in slots, that each epoch should last. NOTE: Currently it is not possible to change the epoch duration after the chain has started. Attempting to do so will brick block production.",
		),
		primitives.NewMetadataModuleConstant(
			"ExpectedBlockTime",
			sc.ToCompact(metadata.PrimitiveTypesU64),
			sc.BytesToSequenceU8(m.constants.ExpectedBlockTime.Bytes()),
			"The expected average block time at which BABE should be creating blocks. Since BABE is probabilistic it is not trivial to figure out what the expected average block time should be based on the slot duration and the security parameter `c` (where `1 - c` represents the probability of a slot being empty).",
		),
		primitives.NewMetadataModuleConstant(
			"MaxAuthorities",
			sc.ToCompact(metadata.PrimitiveTypesU32),
			sc.BytesToSequenceU8(m.constants.MaxAuthorities.Bytes()),
			"Max number of authorities allowed",
		),
	}
}

func (m module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
		Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
			primitives.NewMetadataModuleStorageEntry(
				"Authorities",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBoundedVecAuthority)),
				"Current epoch authorities.",
			),

			primitives.NewMetadataModuleStorageEntry( // TODO fix
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

			primitives.NewMetadataModuleStorageEntry( // TODO fix
				"EpochStart",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesRationalValueU64)),
				"The block numbers when the last and current epoch have started, respectively `N-1` and `N`. NOTE: We track this is in order to annotate the block number when a given pool of entropy was fixed (i.e. it was known to chain observers). Since epochs are defined in slots, which may be skipped, the block numbers may not line up with the slot numbers.",
			),

			primitives.NewMetadataModuleStorageEntry(
				"GenesisSlot",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSlot)),
				"The slot at which the first epoch actually started. This is 0 until the first block of the chain.",
			),

			primitives.NewMetadataModuleStorageEntry( // TODO fix
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

			primitives.NewMetadataModuleStorageEntry( // TODO fix
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

			primitives.NewMetadataModuleStorageEntry( // TODO fix
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
	})
}

func (m module) Metadata() primitives.MetadataModule {
	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesBabeCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(
						metadata.TypesBabeCalls,
						"self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Babe, Runtime>",
					),
				},
				m.index,
				"Call.Babe",
			),
		),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: m.metadataConstants(),
		Error:     sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesBabeErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesBabeErrors),
				},
				m.index,
				"Errors.Babe",
			),
		),
		Index: m.index,
	}

	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}
