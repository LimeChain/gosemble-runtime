package balances

import primitives "github.com/LimeChain/gosemble/primitives/types"

func callTransferAllowDeathWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(47297000, 3593).
		SaturatingAdd(dbWeight.Reads(1)).
		SaturatingAdd(dbWeight.Writes(1))
}
