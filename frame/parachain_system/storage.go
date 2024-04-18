package parachain_system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/parachain"
)

var (
	// module prefix
	keyParachainSystem           = []byte("ParachainSystem")
	keyCustomValidationHeadData  = []byte("CustomValidationHeadData")
	keyNewValidationCode         = []byte("NewValidationCode")
	keyHrmpOutboundMessages      = []byte("HrmpOutboundMessages")
	keyHrmpWatermark             = []byte("HrmpWatermark")
	keyProcessedDownwardMessages = []byte("ProcessedDownwardMessages")
	keyUpwardMessages            = []byte("UpwardMessages")
)

type storage struct {
	// NewValidationCode is the validation code, which is set by the parachain and is to be communicated to the collator
	// relay chain.
	NewValidationCode         support.StorageValue[sc.Sequence[sc.U8]]
	ProcessedDownwardMessages support.StorageValue[sc.U32]
	HrmpWatermark             support.StorageValue[sc.U32]
	HrmpOutboundMessages      support.StorageValue[sc.Sequence[parachain.OutboundHrmpMessage]]
	UpwardMessages            support.StorageValue[sc.Sequence[parachain.UpwardMessage]]
	CustomValidationHeadData  support.StorageValue[sc.Option[sc.Sequence[sc.U8]]]
}

func newStorage() *storage {
	return &storage{
		NewValidationCode:         support.NewHashStorageValue(keyParachainSystem, keyNewValidationCode, sc.DecodeSequence[sc.U8]),
		ProcessedDownwardMessages: support.NewHashStorageValue(keyParachainSystem, keyProcessedDownwardMessages, sc.DecodeU32),
		HrmpWatermark:             support.NewHashStorageValue(keyParachainSystem, keyHrmpWatermark, sc.DecodeU32),
		HrmpOutboundMessages:      support.NewHashStorageValue(keyParachainSystem, keyHrmpOutboundMessages, parachain.DecodeOutboundHrmpMessages),
		UpwardMessages:            support.NewHashStorageValue(keyParachainSystem, keyUpwardMessages, parachain.DecodeUpwardMessages),
		CustomValidationHeadData:  support.NewHashStorageValue(keyParachainSystem, keyCustomValidationHeadData, sc.DecodeOption[sc.Sequence[sc.U8]]),
	}
}
