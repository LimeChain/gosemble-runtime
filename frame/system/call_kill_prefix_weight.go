// THIS FILE WAS GENERATED USING GOSEMBLE BENCHMARKING PACKAGE
// DATE: `2024-06-12 10:03:34.391686 +0300 EEST m=+21.447280042`, STEPS: `50`, REPEAT: `20`, DBCACHE: `1024`, HEAPPAGES: `4096`, HOSTNAME: `Rados-MBP.lan`, CPU: `Apple M1 Pro(8 cores, 3228 mhz)`, GC: ``, TINYGO VERSION: ``, TARGET: ``

// Summary:
// BaseExtrinsicTime: 115594308, BaseReads: 0, BaseWrites: 0, SlopesExtrinsicTime: [3436991], SlopesReads: [1], SlopesWrites: [1], MinExtrinsicTime: 86950, MinReads: 1, MinWrites: 1

package system

import (sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func callKillPrefixWeight(dbWeight primitives.RuntimeDbWeight, size sc.U64) primitives.Weight {
	return primitives.WeightFromParts(115594308, 0).
			SaturatingAdd(primitives.WeightFromParts(3436991, 0).SaturatingMul(size)).
		SaturatingAdd(dbWeight.Reads(0)).
			SaturatingAdd(dbWeight.Reads(1).SaturatingMul(size)).
		SaturatingAdd(dbWeight.Writes(0)).
			SaturatingAdd(dbWeight.Writes(1).SaturatingMul(size))
}
