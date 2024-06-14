package parachain_system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

func enqueueInboundDownwardMessagesWeight(n sc.U64, dbWeight types.RuntimeDbWeight) types.Weight {
	return types.WeightFromParts(1_735_000, 8013).
		SaturatingAdd(types.WeightFromParts(25_300_108, 0).
			SaturatingMul(n),
		).
		SaturatingAdd(dbWeight.Reads(4)).
		SaturatingAdd(dbWeight.Writes(4))
}
