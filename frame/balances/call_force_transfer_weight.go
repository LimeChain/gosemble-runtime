package balances

import primitives "github.com/LimeChain/gosemble/primitives/types"

func callForceTransferWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(1773600000, 0).
		SaturatingAdd(dbWeight.Reads(2)).
		SaturatingAdd(dbWeight.Writes(2))
}
