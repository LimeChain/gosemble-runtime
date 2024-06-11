package balances

import primitives "github.com/LimeChain/gosemble/primitives/types"

func callTransferAllWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(2083900000, 0).
		SaturatingAdd(dbWeight.Reads(1)).
		SaturatingAdd(dbWeight.Writes(1))
}
