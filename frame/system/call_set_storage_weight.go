// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-06-12 10:05:30.644281 +0300 EEST m=+137.700031542`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MBP.lan`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 0, BaseReads: 0, BaseWrites: 0, SlopesExtrinsicTime: [7367302], SlopesReads: [0], SlopesWrites: [1], MinExtrinsicTime: 76200, MinReads: 0, MinWrites: 0

package system

import (sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callSetStorageWeight(dbWeight primitives.RuntimeDbWeight, size sc.U64) primitives.Weight {
	return primitives.WeightFromParts(0, 0).
			SaturatingAdd(primitives.WeightFromParts(7367302, 0).SaturatingMul(size)).
		SaturatingAdd(dbWeight.Reads(0)).
		SaturatingAdd(dbWeight.Writes(0)).
			SaturatingAdd(dbWeight.Writes(1).SaturatingMul(size))
}
