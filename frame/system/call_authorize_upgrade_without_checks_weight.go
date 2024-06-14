// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-06-12 10:03:20.982069 +0300 EEST m=+8.037665292`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MBP.lan`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 208950000, BaseReads: 0, BaseWrites: 1, SlopesExtrinsicTime: [], SlopesReads: [], SlopesWrites: [], MinExtrinsicTime: 208950, MinReads: 0, MinWrites: 1

package system

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callAuthorizeUpgradeWithoutChecksWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(208950000, 0).
		SaturatingAdd(dbWeight.Reads(0)).
		SaturatingAdd(dbWeight.Writes(1))
}
