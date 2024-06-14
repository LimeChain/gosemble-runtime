package parachain_system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	optionNoneIdx sc.U8 = iota
	optionSomeIdx
)

func (m module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithPath(
			metadata.TypesHrmpChannelUpdate,
			"cumulus pallets parachain_system HrmpChannelUpdate",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "parachain_system", "HrmpChannelUpdate"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "msg_count", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "total_bytes", "U32"),
			}),
		),
		primitives.NewMetadataType(metadata.TypesTupleU32HrmpChannelUpdate, "(u32, HrmpChannelUpdate)",
			primitives.NewMetadataTypeDefinitionTuple(sc.Sequence[sc.Compact]{
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.ToCompact(metadata.TypesHrmpChannelUpdate),
			})),
		primitives.NewMetadataType(metadata.TypesSequenceTupleU32HrmpChannelUpdate,
			"[](u32, Vec<InboundHrmpMessages>)",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleU32HrmpChannelUpdate))),

		primitives.NewMetadataTypeWithParams(metadata.TypesHrmpOutgoing,
			"cumulus primitives HrmpOutgoing",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "HrmpOutgoing"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceTupleU32HrmpChannelUpdate),
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU32, "K"),
				primitives.NewMetadataTypeParameter(metadata.TypesHrmpChannelUpdate, "V"),
			},
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesUsedBandwidth,
			"cumulus pallets parachain_system UsedBandwidth",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "parachain_system", "UsedBandwidth"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "ump_msg_count", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "ump_total_bytes", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesHrmpOutgoing, "hrmp_outgoing", "BTreeMap<ParaId, HrmpChannelUpdate>"),
			}),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesUpgradeGoAhead,
			"UpgradeGoAhead",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "parachain_system", "UpgradeGoAhead"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Abort",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						parachain.UpgradeGoAheadAbort,
						"UpgradeGoAhead.Abort"),
					primitives.NewMetadataDefinitionVariant(
						"GoAhead",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						parachain.UpgradeGoAheadGoAhead,
						"UpgradeGoAhead.GoAhead"),
				})),

		primitives.NewMetadataTypeWithParam(
			metadata.TypesOptionUpgradeGoAhead,
			"Option<UpgradeGoAhead>",
			sc.Sequence[sc.Str]{"Option"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"None",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						optionNoneIdx,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Some",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesUpgradeGoAhead),
						},
						optionSomeIdx,
						""),
				}),
			primitives.NewMetadataTypeParameter(metadata.TypesUpgradeGoAhead, "T"),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesUpgradeRestriction,
			"UpgradeRestriction",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "parachain_system", "UpgradeRestriction"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Present",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						parachain.UpgradeRestrictionSignalPresent,
						"UpgradeRestriction.Present"),
				})),

		primitives.NewMetadataTypeWithParam(
			metadata.TypesOptionUpgradeRestriction,
			"Option<UpgradeRestriction>",
			sc.Sequence[sc.Str]{"Option"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"None",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						optionNoneIdx,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Some",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesUpgradeRestriction),
						},
						optionSomeIdx,
						""),
				}),
			primitives.NewMetadataTypeParameter(metadata.TypesUpgradeRestriction, "T"),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesAncestor,
			"cumulus pallets parachain_system Ancestor",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "parachain_system", "Ancestor"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesUsedBandwidth, "used_bandwidth", "UsedBandwidth"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionH256, "ump_total_bytes", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionUpgradeGoAhead, "consumed_go_ahead_signal", "Option<UpgradeGoAhead>"),
			}),
		),

		primitives.NewMetadataType(metadata.TypesSequenceAncestor,
			"[]Ancestor",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesAncestor))),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesSegmentTracker,
			"cumulus pallets parachain_system SegmentTracker",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "parachain_system", "SegmentTracker"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesUsedBandwidth, "used_bandwidth", "UsedBandwidth"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionU32, "hrmp_watermark", "Option<u32>"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionUpgradeGoAhead, "consumed_go_ahead_signal", "Option<UpgradeGoAhead>"),
			}),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesRelayDispatchQueueRemainingCapacity,
			"cumulus primitives parachain_system RelayDispatchQueueRemainingCapacity",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "parachain_system", "RelayDispatchQueueRemainingCapacity"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "remaining_count", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "remaining_size", "U32"),
			}),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesAbridgedHrmpChannel,
			"cumulus primitives parachain_system AbridgedHrmpChannel",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "parachain_system", "AbridgedHrmpChannel"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_capacity", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_total_size", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_message_size", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "msg_count", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "total_size", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionH256, "mqc_head", "Option<H256>"),
			}),
		),

		primitives.NewMetadataType(metadata.TypesTupleU32AbridgedHrmpChannel, "(u32, AbridgedHrmpChannel)",
			primitives.NewMetadataTypeDefinitionTuple(sc.Sequence[sc.Compact]{
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.ToCompact(metadata.TypesAbridgedHrmpChannel),
			})),
		primitives.NewMetadataType(metadata.TypesSequenceTupleU32AbridgedHrmpChannel,
			"[](u32, AbridgedHrmpChannel)",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleU32AbridgedHrmpChannel))),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesMessagingStateSnapshot,
			"cumulus primitives parachain_system MessagingStateSnapshot",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "parachain_system", "MessagingStateSnapshot"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesH256, "dmq_mqc_head", "H256"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesRelayDispatchQueueRemainingCapacity, "relay_dispatch_queue_remaining_capacity", "RelayDispatchQueueRemainingCapacity"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceTupleU32AbridgedHrmpChannel, "ingress_channels", "Vec<(ParaId, AbridgedHrmpChannel)>"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceTupleU32AbridgedHrmpChannel, "egress_channels", "Vec<(ParaId, AbridgedHrmpChannel)>"),
			}),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesAsyncBackingParams,
			"cumulus primitives parachain_system AsyncBackingParams",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "parachain_system", "AsyncBackingParams"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_candidate_depth", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "allowed_ancestry_len", "U32"),
			}),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesAbridgedHostConfiguration,
			"cumulus primitives parachain_system AbridgedHostConfiguration",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "parachain_system", "AbridgedHostConfiguration"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_code_size", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_head_data_size", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_upward_queue_count", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_upward_queue_size", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_upward_message_size", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_upward_message_num_per_candidate", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "hrmp_max_message_num_per_candidate", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "validation_upgrade_cooldown", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "validation_upgrade_delay", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAsyncBackingParams, "async_backing_params", "AsyncBackingParams"),
			}),
		),

		primitives.NewMetadataTypeWithParam(metadata.TypesOptionXcmHash, "Option<XcmHash>", sc.Sequence[sc.Str]{"Option"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant(
					"None",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{},
					optionNoneIdx,
					""),
				primitives.NewMetadataDefinitionVariant(
					"Some",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionField(metadata.TypesFixedSequence32U8),
					},
					optionSomeIdx,
					""),
			}),
			primitives.NewMetadataTypeParameter(metadata.TypesFixedSequence32U8, "T"),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesPersistedValidationData,
			"primitives PersistedValidationData",
			sc.Sequence[sc.Str]{"primitives", "PersistedValidationData"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceU8, "parent_head", "ParentHead"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "relay_parent_number", "U32"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesH256, "relay_parent_storage_root", "H256"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "max_pov_size", "U32"),
			}),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesInboundDownwardMessage,
			"cumulus primitives InboundDownwardMessage",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "InboundDownwardMessage"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "sent_at", "BlockNumber"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceU8, "msg", "DownwardMessage"),
			}),
		),
		primitives.NewMetadataType(metadata.TypesSequenceInboundDownwardMessages,
			"[]InboundDownwardMessage",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesInboundDownwardMessage)),
		),
		primitives.NewMetadataTypeWithPath(
			metadata.TypesInboundHrmpMessage,
			"cumulus primitives InboundHrmpMessage",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "InboundHrmpMessage"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "sent_at", "BlockNumber"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceU8, "data", "Vec<u8>"),
			}),
		),
		primitives.NewMetadataType(metadata.TypesSequenceInboundHrmpMessages,
			"[]InboundHrmpMessages",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesInboundHrmpMessage)),
		),

		primitives.NewMetadataType(metadata.TypesTupleU32SequenceInboundHrmpMessages, "(u32, Vec<InboundHrmpMessages>)",
			primitives.NewMetadataTypeDefinitionTuple(sc.Sequence[sc.Compact]{
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.ToCompact(metadata.TypesSequenceInboundHrmpMessages),
			})),
		primitives.NewMetadataType(metadata.TypesSequenceTupleU32SequenceInboundHrmpMessages,
			"[](u32, Vec<InboundHrmpMessages>)",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleU32SequenceInboundHrmpMessages))),

		primitives.NewMetadataTypeWithParams(metadata.TypesHorizontalMessages,
			"cumulus primitives HorizontalMessages",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "HorizontalMessages"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceTupleU32SequenceInboundHrmpMessages),
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU32, "K"),
				primitives.NewMetadataTypeParameter(metadata.TypesSequenceInboundHrmpMessages, "V"),
			},
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesParachainInherentData,
			"cumulus primitives parachain_inherent ParachainInherentData",
			sc.Sequence[sc.Str]{"cumulus", "primitives", "parachain_inherent", "ParachainInherentData"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesPersistedValidationData, "validation_data", "PersistedValidationData"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceSequenceU8, "relay_chain_state", "StorageProof"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceInboundDownwardMessages, "downward_messages", "Vec<InboundDownwardMessage>"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesHorizontalMessages, "horizontal_messages", "BTreeMap<ParaId,Vec<InboundHrmpMessage>>"),
			}),
		),

		primitives.NewMetadataTypeWithPath(
			metadata.TypesParachainSystemEvents,
			"cumulus pallets ParachainSystem Event",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "ParachainSystem", "Event"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"ValidationFunctionStored",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventValidationFunctionStored,
						"Events.ValidationFunctionStored"),
					primitives.NewMetadataDefinitionVariant(
						"ValidationFunctionApplied",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "relay_chain_block_num", "U32"),
						},
						EventValidationFunctionApplied,
						"Events.ValidationFunctionApplied"),
					primitives.NewMetadataDefinitionVariant(
						"ValidationFunctionDiscarded",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventValidationFunctionDiscarded,
						"Events.ValidationFunctionDiscarded"),
					primitives.NewMetadataDefinitionVariant(
						"DownwardMessagesReceived",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU32, "count", "U32"),
						},
						EventDownwardMessagesReceived,
						"Events.DownwardMessagesReceived"),
					primitives.NewMetadataDefinitionVariant(
						"DownwardMessagesProcessed",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesWeight, "weight_used", "Weight"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesH256, "dmq_head", "H256"),
						},
						EventDownwardMessagesProcessed,
						"Events.DownwardMessagesProcessed"),
					primitives.NewMetadataDefinitionVariant(
						"UpwardMessageSent",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionXcmHash, "message_hash", "Option<XcmHash>"),
						},
						EventUpwardMessageSent,
						"Events.UpwardMessageSent"),
				})),
		primitives.NewMetadataTypeWithParam(metadata.TypesParachainSystemCalls,
			"ParachainSystem calls",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "parachain_system", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"setValidationData",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesParachainInherentData, "data", "ParachainInherentData"),
						},
						FunctionSetValidationData,
						"Set the current validation data. This should be invoked exactly once per block. "+
							"It will panic at the finalisation if the call was not invoked. "+
							"The dispatch origin for this call must be `Inherent`. "+
							"As a side effect, this function upgrades the current validation function if the appropriate time has come."),
					primitives.NewMetadataDefinitionVariant(
						"sudo_unchecked_weight",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceU8, "message", "UpwardMessage"),
						},
						functionSudoSendUpwardMessage,
						"Sends an upward message.",
					),
				}),
			primitives.NewMetadataEmptyTypeParameter("T")),

		primitives.NewMetadataTypeWithParams(metadata.TypesParachainSystemErrors,
			"parachain_system pallet Error",
			sc.Sequence[sc.Str]{"cumulus", "pallets", "parachain_system", "Error"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"OverlappingUpgrades",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorOverlappingUpgrades,
						"Attempt to upgrade validation function while existing upgrade pending.",
					),
					primitives.NewMetadataDefinitionVariant(
						"ProhibitedByPolkadot",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorProhibitedByPolkadot,
						"Polkadot currently prohibits this parachain from upgrading its validation function.",
					),
					primitives.NewMetadataDefinitionVariant(
						"TooBig",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorTooBig,
						"The supplied validation function has compiled into a blob larger than Polkadot is willing to run.",
					),
					primitives.NewMetadataDefinitionVariant(
						"ValidationDataNotAvailable",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorValidationDataNotAvailable,
						"The inherent which supplies the validation data did not run this block.",
					),
					primitives.NewMetadataDefinitionVariant(
						"HostConfigurationNotAvailable",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorHostConfigurationNotAvailable,
						"The inherent which supplies the host configuration did not run this block.",
					),
					primitives.NewMetadataDefinitionVariant(
						"NotScheduled",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorNotScheduled,
						"No validation function upgrade is currently scheduled.",
					),
					primitives.NewMetadataDefinitionVariant(
						"NothingAuthorized",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorNothingAuthorized,
						"No code upgrade has been authorized.",
					),
					primitives.NewMetadataDefinitionVariant(
						"Unauthorized",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorUnauthorized,
						"The given code upgrade has not been authorized.",
					),
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			}),
	}
}
