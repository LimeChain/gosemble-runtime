// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-04-29 22:01:01.470495 +0300 EEST m=+1.966673542`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MBP.lan`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 96100000, BaseReads: 0, BaseWrites: 1, SlopesExtrinsicTime: [], SlopesReads: [], SlopesWrites: [], MinExtrinsicTime: 96100, MinReads: 0, MinWrites: 1

package grandpa

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callNoteStalledWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(96100000, 0).
		SaturatingAdd(dbWeight.Reads(0)).
		SaturatingAdd(dbWeight.Writes(1))
}
