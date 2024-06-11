package babe

import sc "github.com/LimeChain/goscale"

type consts struct {
	EpochDuration     sc.U64
	ExpectedBlockTime sc.U64
	MaxAuthorities    sc.U32
}

func newConstants(epochDuration sc.U64, expectedBlockTime sc.U64, maxAuthorities sc.U32) *consts {
	return &consts{
		EpochDuration:     epochDuration,
		ExpectedBlockTime: expectedBlockTime,
		MaxAuthorities:    maxAuthorities,
	}
}
