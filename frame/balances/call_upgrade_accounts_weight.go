// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-04-02 04:45:52.24591 +0300 EEST m=+2.782190168`, STEPS: `2`, REPEAT: `1`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `MacBook-Pro.local`, CPU: `Apple M2 Pro(10 cores, 3504 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 19211211, BaseReads: 0, BaseWrites: 0, SlopesExtrinsicTime: [663788788], SlopesReads: [1], SlopesWrites: [1], MinExtrinsicTime: 683000, MinReads: 1, MinWrites: 1

package balances

import (sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callUpgradeAccountsWeight(dbWeight primitives.RuntimeDbWeight, size sc.U64) primitives.Weight {
	return primitives.WeightFromParts(19211211, 0).
			SaturatingAdd(primitives.WeightFromParts(663788788, 0).SaturatingMul(size)).
		SaturatingAdd(dbWeight.Reads(0)).
			SaturatingAdd(dbWeight.Reads(1).SaturatingMul(size)).
		SaturatingAdd(dbWeight.Writes(0)).
			SaturatingAdd(dbWeight.Writes(1).SaturatingMul(size))
}
