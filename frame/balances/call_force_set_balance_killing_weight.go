package balances

import primitives "github.com/LimeChain/gosemble/primitives/types"

func callForceSetBalanceKillingWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(947950000, 0).
		SaturatingAdd(dbWeight.Reads(2)).
		SaturatingAdd(dbWeight.Writes(2))
}
