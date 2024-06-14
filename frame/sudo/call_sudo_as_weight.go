// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-06-12 10:03:17.132353 +0300 EEST m=+4.187953917`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MBP.lan`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 269250000, BaseReads: 1, BaseWrites: 0, SlopesExtrinsicTime: [], SlopesReads: [], SlopesWrites: [], MinExtrinsicTime: 269250, MinReads: 1, MinWrites: 0

package sudo

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callSudoAsWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(269250000, 0).
		SaturatingAdd(dbWeight.Reads(1)).
		SaturatingAdd(dbWeight.Writes(0))
}
