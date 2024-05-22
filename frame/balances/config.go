package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage            io.Storage
	DbWeight           primitives.RuntimeDbWeight
	MaxLocks           sc.U32
	MaxReserves        sc.U32
	ExistentialDeposit sc.U128
	StoredMap          primitives.StoredMap
}

func NewConfig(storage io.Storage, dbWeight primitives.RuntimeDbWeight, maxLocks sc.U32, maxReserves sc.U32, existentialDeposit sc.U128, storedMap primitives.StoredMap) *Config {
	return &Config{
		Storage:            storage,
		DbWeight:           dbWeight,
		MaxLocks:           maxLocks,
		MaxReserves:        maxReserves,
		ExistentialDeposit: existentialDeposit,
		StoredMap:          storedMap,
	}
}
