package balances

import primitives "github.com/LimeChain/gosemble/primitives/types"

func callForceSetBalanceCreatingWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(705450000, 0).
		SaturatingAdd(dbWeight.Reads(2)).
		SaturatingAdd(dbWeight.Writes(2))
}
