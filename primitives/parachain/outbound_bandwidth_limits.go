package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type OutboundBandwidthLimits struct {
	UmpMessagesRemaining sc.U32
	UmpBytesRemaining    sc.U32
	HrmpOutgoing         sc.Dictionary[sc.U32, HrmpOutboundLimits]
}

func NewOutboundBandwidthLimitsFromMessagingStateSnapshot(mss MessagingStateSnapshot) OutboundBandwidthLimits {
	var hrmpOutgoing sc.Dictionary[sc.U32, HrmpOutboundLimits]
	for _, abridgedHrmpChannel := range mss.EgressChannels {
		bytesRemaining := sc.SaturatingSubU64(sc.U64(abridgedHrmpChannel.AbridgedHRMPChannel.MaxTotalSize), sc.U64(abridgedHrmpChannel.AbridgedHRMPChannel.TotalSize))
		messagesRemaining := sc.SaturatingSubU64(sc.U64(abridgedHrmpChannel.AbridgedHRMPChannel.MaxCapacity), sc.U64(abridgedHrmpChannel.AbridgedHRMPChannel.MsgCount))

		hrmpOutgoing[abridgedHrmpChannel.ParachainId] = HrmpOutboundLimits{
			BytesRemaining:    sc.U32(bytesRemaining),
			MessagesRemaining: sc.U32(messagesRemaining),
		}
	}

	return OutboundBandwidthLimits{
		UmpMessagesRemaining: mss.RelayDispatchQueueRemainingCapacity.RemainingCount,
		UmpBytesRemaining:    mss.RelayDispatchQueueRemainingCapacity.RemainingSize,
	}
}

func (obl OutboundBandwidthLimits) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, obl.UmpBytesRemaining, obl.UmpBytesRemaining, obl.HrmpOutgoing)
}

func (obl OutboundBandwidthLimits) Bytes() []byte {
	return sc.EncodedBytes(obl)
}
