package grandpa

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func (m module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithParams(
			metadata.TypesGrandpaErrors,
			"The `Error` enum of this pallet.",
			sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Error"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant("PauseFailed", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorPauseFailed, ""),
					primitives.NewMetadataDefinitionVariant("ResumeFailed", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorResumeFailed, ""),
					primitives.NewMetadataDefinitionVariant("ChangePending", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorChangePending, ""),
					primitives.NewMetadataDefinitionVariant("TooSoon", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorTooSoon, ""),
					primitives.NewMetadataDefinitionVariant("InvalidKeyOwnershipProof", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorInvalidKeyOwnershipProof, ""),
					primitives.NewMetadataDefinitionVariant("InvalidEquivocationProof", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorInvalidEquivocationProof, ""),
					primitives.NewMetadataDefinitionVariant("DuplicateOffenceReport", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ErrorDuplicateOffenceReport, ""),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
			},
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesGrandpaAppPublic,
			"sp_consensus_grandpa app Public",
			sc.Sequence[sc.Str]{"sp_consensus_grandpa", "app", "Public"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesEd25519PubKey),
				},
			),
		),

		primitives.NewMetadataType(
			metadata.TypesTupleGrandpaAppPublicU64,
			"(GrandpaAppPublic, U64)",
			primitives.NewMetadataTypeDefinitionTuple(
				sc.Sequence[sc.Compact]{
					sc.ToCompact(metadata.TypesGrandpaAppPublic), sc.ToCompact(metadata.PrimitiveTypesU64),
				},
			),
		),

		primitives.NewMetadataType(
			metadata.TypesSequenceTupleGrandpaAppPublic,
			"[]byte (GrandpaAppPublic, U64)",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleGrandpaAppPublicU64)),
		),

		primitives.NewMetadataTypeWithParams(
			metadata.TypesGrandpaStoredPendingChange,
			"StoredPendingChange",
			sc.Sequence[sc.Str]{"pallet_grandpa", "StoredPendingChange"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "scheduled_at", "u64"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "delay", "u64"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesBoundedVecAuthority, "next_authorities", "BoundedAuthorityList<Limit>"),
					primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionU64, "forced", "Option<u64>"),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU64, "N"),
				primitives.NewMetadataEmptyTypeParameter("Limit"),
			},
		),

		primitives.NewMetadataTypeWithParams(
			metadata.TypesGrandpaStoredState,
			"StoredState",
			sc.Sequence[sc.Str]{"pallet_grandpa", "StoredState"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant("Live", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, StoredStateLive, ""),
					primitives.NewMetadataDefinitionVariant("PendingPause", sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "scheduled_at", "u64"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "delay", "u64"),
					}, StoredStatePendingPause, ""),
					primitives.NewMetadataDefinitionVariant("Paused", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, StoredStatePaused, ""),
					primitives.NewMetadataDefinitionVariant("PendingResume", sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "scheduled_at", "u64"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU64, "delay", "u64"),
					}, StoredStatePendingResume, ""),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU64, "N"),
			},
		),

		primitives.NewMetadataTypeWithParam(
			metadata.TypesGrandpaCalls,
			"Grandpa calls",
			sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"note_stalled",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU64, "delay"),
							primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU64, "best_finalized_block_number"),
						},
						functionNoteStalledIndex,
						"Note that the current authority set of the GRANDPA finality gadget has stalled. This will trigger a forced authority set change at the beginning of the next session, to be enacted `delay` blocks after that. The `delay` should be high enough to safely assume that the block signalling the forced change will not be re-orged e.g. 1000 blocks. The block production rate (which may be slowed down because of finality lagging) should be taken into account when choosing the `delay`. The GRANDPA voters based on the new authority will start voting on top of `best_finalized_block_number` for new finalized blocks. `best_finalized_block_number` should be the highest of the latest finalized block of all validators of the new authority set. Only callable by root.",
					),
				},
			),
			primitives.NewMetadataEmptyTypeParameter("T"),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesGrandpaEvent,
			"pallet_grandpa pallet Event",
			sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Event"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"NewAuthorities",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceAuthority, "authority_set", "AuthorityList"),
						},
						EventNewAuthorities,
						"New authority set has been applied.",
					),

					primitives.NewMetadataDefinitionVariant(
						"Paused",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventPaused,
						"Current authority set has been paused.",
					),

					primitives.NewMetadataDefinitionVariant(
						"Resumed",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventResumed,
						"Current authority set has been resumed.",
					),
				},
			),
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
				"The current list of authorities.",
			),

			primitives.NewMetadataModuleStorageEntry(
				"CurrentSetId",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU64)),
				"The number of changes (both in terms of keys and underlying economic responsibilities) in the \"set\" of Grandpa validators from genesis.",
			),

			primitives.NewMetadataModuleStorageEntry(
				"Stalled",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesTuple2U64)),
				"`true` if we are currently stalled.",
			),

			primitives.NewMetadataModuleStorageEntry(
				"PendingChange",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesGrandpaStoredPendingChange)),
				"Pending change: (signaled at, scheduled change).",
			),

			primitives.NewMetadataModuleStorageEntry(
				"State",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesGrandpaStoredState)),
				"State of the current authority set.",
			),

			primitives.NewMetadataModuleStorageEntry(
				"SetIdSession",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionMap(
					sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
					sc.ToCompact(metadata.PrimitiveTypesU64),
					sc.ToCompact(metadata.PrimitiveTypesU32),
				),
				"A mapping from grandpa set ID to the index of the *most recent* session for which its members were responsible. This is only used for validating equivocation proofs. An equivocation proof must contains a key-ownership proof for a given session, therefore we need a way to tie together sessions and GRANDPA set ids, i.e. we need to validate that a validator was the owner of a given key on a given session, and what the active set ID was during that session. TWOX-NOTE: `SetId` is not under user control.",
			),

			primitives.NewMetadataModuleStorageEntry(
				"NextForced",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU64)),
				"Next block number where we can force a change.",
			),
		},
	})
}

func (m module) metadataConstants() sc.Sequence[primitives.MetadataModuleConstant] {
	return sc.Sequence[primitives.MetadataModuleConstant]{
		primitives.NewMetadataModuleConstant(
			"MaxAuthorities",
			sc.ToCompact(metadata.PrimitiveTypesU32),
			sc.BytesToSequenceU8(m.constants.MaxAuthorities.Bytes()),
			"Max Authorities in use.",
		),
		primitives.NewMetadataModuleConstant(
			"MaxNominators",
			sc.ToCompact(metadata.PrimitiveTypesU32),
			sc.BytesToSequenceU8(m.constants.MaxNominators.Bytes()),
			"The maximum number of nominators for each validator.",
		),
		primitives.NewMetadataModuleConstant(
			"MaxSetIdSessionEntries",
			sc.ToCompact(metadata.PrimitiveTypesU64),
			sc.BytesToSequenceU8(m.constants.MaxSetIdSessionEntries.Bytes()),
			"The maximum number of entries to keep in the set id to session index mapping. Since the `SetIdSession` map is only used for validating equivocations this value should relate to the bonding duration of whatever staking system is being used (if any). If equivocation handling is not enabled then this value can be zero.",
		),
	}
}

func (m module) Metadata() primitives.MetadataModule {
	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesGrandpaCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(
						metadata.TypesGrandpaCalls,
						"self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Grandpa, Runtime>",
					),
				},
				m.index,
				"Call.Grandpa",
			),
		),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesGrandpaEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesGrandpaEvent, "pallet_grandpa::Event<Runtime>"),
				},
				m.index,
				"Events.Grandpa",
			),
		),
		Constants: m.metadataConstants(),
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesGrandpaErrors),
				},
				m.index,
				"Errors.Grandpa",
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
