package aura_ext

import (
	"github.com/LimeChain/gosemble/primitives/types"
)

type consts struct {
	DbWeight types.RuntimeDbWeight
}

func newConstants(dbWeight types.RuntimeDbWeight) consts {
	return consts{
		dbWeight,
	}
}
