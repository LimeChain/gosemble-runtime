package parachain_system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/parachain"
)

var (
	// module prefix
	keyParachainSystem                   = []byte("ParachainSystem")
	keyUnincludedSegment                 = []byte("UnincludedSegment")
	keyAggregatedUnincludedSegment       = []byte("AggregatedUnincludedSegment")
	keyCustomValidationHeadData          = []byte("CustomValidationHeadData")
	keyPendingValidationCode             = []byte("PendingValidationCode")
	keyNewValidationCode                 = []byte("NewValidationCode")
	keyValidationData                    = []byte("ValidationData")
	keyDidSetValidationCode              = []byte("DidSetValidationCode")
	keyLastRelayChainBlockNumber         = []byte("LastRelayChainBlockNumber")
	keyUpgradeRestrictionSignal          = []byte("UpgradeRestrictionSignal")
	keyUpgradeGoAhead                    = []byte("UpgradeGoAhead")
	keyRelayStateProof                   = []byte("RelayStateProof")
	keyRelevantMessagingState            = []byte("RelevantMessagingState")
	keyHostConfiguration                 = []byte("HostConfiguration")
	keyHrmpOutboundMessages              = []byte("HrmpOutboundMessages")
	keyHrmpWatermark                     = []byte("HrmpWatermark")
	keyProcessedDownwardMessages         = []byte("ProcessedDownwardMessages")
	keyAnnouncedHrmpMessagesPerCandidate = []byte("AnnouncedHrmpMessagesPerCandidate")
	keyUpwardMessages                    = []byte("UpwardMessages")
)

type storage struct {
	UnincludedSegment           support.StorageValue[parachain.UnincludedSegment]
	AggregatedUnincludedSegment support.StorageValue[parachain.SegmentTracker]
	PendingValidationCode       support.StorageValue[sc.Sequence[sc.U8]]
	// NewValidationCode is the validation code, which is set by the parachain and is to be communicated to the collator
	// relay chain.
	NewValidationCode                 support.StorageValue[sc.Sequence[sc.U8]]
	ValidationData                    support.StorageValue[parachain.PersistedValidationData]
	DidSetValidationCode              support.StorageValue[sc.Bool]
	LastRelayChainBlockNumber         support.StorageValue[parachain.RelayChainBlockNumber]
	UpgradeRestrictionSignal          support.StorageValue[sc.Option[sc.U8]]
	UpgradeGoAhead                    support.StorageValue[sc.Option[sc.U8]]
	RelayStateProof                   support.StorageValue[parachain.StorageProof]
	RelevantMessagingState            support.StorageValue[parachain.MessagingStateSnapshot]
	HostConfiguration                 support.StorageValue[parachain.AbridgedHostConfiguration]
	ProcessedDownwardMessages         support.StorageValue[sc.U32]
	HrmpWatermark                     support.StorageValue[sc.U32]
	HrmpOutboundMessages              support.StorageValue[sc.Sequence[parachain.OutboundHrmpMessage]]
	UpwardMessages                    support.StorageValue[sc.Sequence[parachain.UpwardMessage]]
	AnnouncedHrmpMessagesPerCandidate support.StorageValue[sc.U32]
	CustomValidationHeadData          support.StorageValue[sc.Option[sc.Sequence[sc.U8]]]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		UnincludedSegment:                 support.NewHashStorageValue(s, keyParachainSystem, keyUnincludedSegment, parachain.DecodeUnincludedSegment),
		AggregatedUnincludedSegment:       support.NewHashStorageValue(s, keyParachainSystem, keyAggregatedUnincludedSegment, parachain.DecodeSegmentTracker),
		PendingValidationCode:             support.NewHashStorageValue(s, keyParachainSystem, keyPendingValidationCode, sc.DecodeSequence[sc.U8]),
		NewValidationCode:                 support.NewHashStorageValue(s, keyParachainSystem, keyNewValidationCode, sc.DecodeSequence[sc.U8]),
		ValidationData:                    support.NewHashStorageValue(s, keyParachainSystem, keyValidationData, parachain.DecodePersistedValidationData),
		DidSetValidationCode:              support.NewHashStorageValue(s, keyParachainSystem, keyDidSetValidationCode, sc.DecodeBool),
		LastRelayChainBlockNumber:         support.NewHashStorageValue(s, keyParachainSystem, keyLastRelayChainBlockNumber, sc.DecodeU32),
		UpgradeRestrictionSignal:          support.NewHashStorageValue(s, keyParachainSystem, keyUpgradeRestrictionSignal, parachain.DecodeOptionUpgradeRestrictionSignal),
		UpgradeGoAhead:                    support.NewHashStorageValue(s, keyParachainSystem, keyUpgradeGoAhead, parachain.DecodeOptionUpgradeGoAhead),
		RelayStateProof:                   support.NewHashStorageValue(s, keyParachainSystem, keyRelayStateProof, parachain.DecodeStorageProof),
		RelevantMessagingState:            support.NewHashStorageValue(s, keyParachainSystem, keyRelevantMessagingState, parachain.DecodeMessagingStateSnapshot),
		HostConfiguration:                 support.NewHashStorageValue(s, keyParachainSystem, keyHostConfiguration, parachain.DecodeAbridgeHostConfiguration),
		ProcessedDownwardMessages:         support.NewHashStorageValue(s, keyParachainSystem, keyProcessedDownwardMessages, sc.DecodeU32),
		HrmpWatermark:                     support.NewHashStorageValue(s, keyParachainSystem, keyHrmpWatermark, sc.DecodeU32),
		HrmpOutboundMessages:              support.NewHashStorageValue(s, keyParachainSystem, keyHrmpOutboundMessages, parachain.DecodeOutboundHrmpMessages),
		UpwardMessages:                    support.NewHashStorageValue(s, keyParachainSystem, keyUpwardMessages, parachain.DecodeUpwardMessages),
		AnnouncedHrmpMessagesPerCandidate: support.NewHashStorageValue(s, keyParachainSystem, keyAnnouncedHrmpMessagesPerCandidate, sc.DecodeU32),
		CustomValidationHeadData:          support.NewHashStorageValue(s, keyParachainSystem, keyCustomValidationHeadData, sc.DecodeOption[sc.Sequence[sc.U8]]),
	}
}
