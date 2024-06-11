package grandpa

import sc "github.com/LimeChain/goscale"

type consts struct {
	MaxAuthorities         sc.U32
	MaxNominators          sc.U32
	MaxSetIdSessionEntries sc.U64
}

func newConstants(maxAuthorities sc.U32, maxNominators sc.U32, maxSetIdSessionEntries sc.U64) *consts {
	return &consts{
		MaxAuthorities:         maxAuthorities,
		MaxNominators:          maxNominators,
		MaxSetIdSessionEntries: maxSetIdSessionEntries,
	}
}
