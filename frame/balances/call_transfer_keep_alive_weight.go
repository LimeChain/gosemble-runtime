package balances

import primitives "github.com/LimeChain/gosemble/primitives/types"

func callTransferKeepAliveWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(1782050000, 0).
		SaturatingAdd(dbWeight.Reads(1)).
		SaturatingAdd(dbWeight.Writes(1))
}
