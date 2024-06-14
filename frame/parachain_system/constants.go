package parachain_system

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const thresholdFactor = 2

// TODO: FixedU128
const exponentialFeeBase = 105

// TODO: FixedU128
const messageSizeFeeBase = 1000

type consts struct {
	DbWeight primitives.RuntimeDbWeight
}

func newConstants(dbWeight primitives.RuntimeDbWeight) consts {
	return consts{
		DbWeight: dbWeight,
	}
}
