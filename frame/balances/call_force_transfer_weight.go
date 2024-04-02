// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-04-02 04:50:15.881078 +0300 EEST m=+2.443391084`, STEPS: `2`, REPEAT: `1`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `MacBook-Pro.local`, CPU: `Apple M2 Pro(10 cores, 3504 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 2061000000, BaseReads: 3, BaseWrites: 3, SlopesExtrinsicTime: [], SlopesReads: [], SlopesWrites: [], MinExtrinsicTime: 1911000, MinReads: 3, MinWrites: 3

package balances

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callForceTransferWeight(dbWeight primitives.RuntimeDbWeight) primitives.Weight {
	return primitives.WeightFromParts(2061000000, 0).
		SaturatingAdd(dbWeight.Reads(3)).
		SaturatingAdd(dbWeight.Writes(3))
}
