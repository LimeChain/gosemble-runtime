// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-06-12 10:04:26.364946 +0300 EEST m=+73.420601751`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MBP.lan`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 92589999, BaseReads: 0, BaseWrites: 0, SlopesExtrinsicTime: [0], SlopesReads: [0], SlopesWrites: [0], MinExtrinsicTime: 72200, MinReads: 0, MinWrites: 0

package system

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callRemarkWeight(dbWeight primitives.RuntimeDbWeight, size sc.U64) primitives.Weight {
	return primitives.WeightFromParts(92589999, 0).
		SaturatingAdd(primitives.WeightFromParts(10, 0).SaturatingMul(size)).
		SaturatingAdd(dbWeight.Reads(0)).
		SaturatingAdd(dbWeight.Writes(0))
}
