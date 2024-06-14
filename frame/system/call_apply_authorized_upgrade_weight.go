// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-06-12 10:03:20.509067 +0300 EEST m=+7.564663334`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MBP.lan`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 116793550000, BaseReads: 2, BaseWrites: 3, SlopesExtrinsicTime: [], SlopesReads: [], SlopesWrites: [], MinExtrinsicTime: 116793550, MinReads: 2, MinWrites: 3

package system

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callApplyAuthorizedUpgradeWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(116793550000, 0).
		SaturatingAdd(dbWeight.Reads(2)).
		SaturatingAdd(dbWeight.Writes(3))
}
