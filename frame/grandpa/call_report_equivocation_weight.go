// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-04-17 13:16:06.741721 +0300 EEST m=+0.272383126`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MacBook-Pro.local`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 113050000, BaseReads: 0, BaseWrites: 1, SlopesExtrinsicTime: [], SlopesReads: [], SlopesWrites: [], MinExtrinsicTime: 113050, MinReads: 0, MinWrites: 1

package grandpa

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callReportEquivocationWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(0, 0).
		SaturatingAdd(dbWeight.Reads(0)).
		SaturatingAdd(dbWeight.Writes(0))
}
