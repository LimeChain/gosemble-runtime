package parachain_system

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type consts struct {
	DbWeight primitives.RuntimeDbWeight
}

func newConstants(dbWeight primitives.RuntimeDbWeight) consts {
	return consts{
		DbWeight: dbWeight,
	}
}
