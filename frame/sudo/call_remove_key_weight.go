// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-04-10 12:57:43.946518 +0300 EEST m=+0.294647501`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: ``, CPU: `Apple M2 Pro(10 cores, 3504 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 536700000, BaseReads: 1, BaseWrites: 1, SlopesExtrinsicTime: [], SlopesReads: [], SlopesWrites: [], MinExtrinsicTime: 536700, MinReads: 1, MinWrites: 1

package sudo

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callRemoveKeyWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(536700000, 0).
		SaturatingAdd(dbWeight.Reads(1)).
		SaturatingAdd(dbWeight.Writes(1))
}
