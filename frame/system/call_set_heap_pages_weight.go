// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-06-12 10:05:14.698223 +0300 EEST m=+121.753949751`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MBP.lan`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 108700000, BaseReads: 1, BaseWrites: 2, SlopesExtrinsicTime: [], SlopesReads: [], SlopesWrites: [], MinExtrinsicTime: 108700, MinReads: 1, MinWrites: 2

package system

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callSetHeapPagesWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(108700000, 0).
		SaturatingAdd(dbWeight.Reads(1)).
		SaturatingAdd(dbWeight.Writes(2))
}
