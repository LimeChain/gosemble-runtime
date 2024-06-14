package balances

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callUpgradeAccountsWeight(dbWeight primitives.RuntimeDbWeight, length sc.U64) primitives.Weight {
	return primitives.WeightFromParts(16118000, 990).
		SaturatingAdd(primitives.WeightFromParts(13327660, 0).SaturatingMul(length)).
		SaturatingAdd(dbWeight.Reads(1).SaturatingMul(length)).
		SaturatingAdd(dbWeight.Writes(1).SaturatingMul(length)).
		SaturatingAdd(primitives.WeightFromParts(0, 2603).SaturatingMul(length))
}
