package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"

	// "github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyBalances         = []byte("Balances")
	keyTotalIssuance    = []byte("TotalIssuance")
	keyInactiveIssuance = []byte("InactiveIssuance")
	// keyLocks            = []byte("Locks")
	// keyReserves         = []byte("Reserves")
	// keyHolds            = []byte("Holds")
	// keyFreezes          = []byte("Freezes")
)

type storage struct {
	TotalIssuance    support.StorageValue[primitives.Balance]
	InactiveIssuance support.StorageValue[primitives.Balance]
	// Locks            support.StorageMap[primitives.AccountId, primitives.Balance]
	// Reserves         support.StorageMap[primitives.AccountId, primitives.Balance]
	// Holds            support.StorageMap[primitives.AccountId, primitives.Balance]
	// Freezes          support.StorageMap[primitives.AccountId, primitives.Balance]
}

func newStorage() *storage {
	// hashing := io.NewHashing()

	return &storage{
		TotalIssuance:    support.NewHashStorageValue(keyBalances, keyTotalIssuance, sc.DecodeU128),
		InactiveIssuance: support.NewHashStorageValue(keyBalances, keyInactiveIssuance, sc.DecodeU128),
		// Locks:            support.NewHashStorageMap[primitives.AccountId](keyBalances, keyLocks, hashing.Blake128, sc.DecodeU128),
		// Reserves:         support.NewHashStorageMap[primitives.AccountId](keyBalances, keyReserves, hashing.Blake128, sc.DecodeU128),
		// Holds:            support.NewHashStorageMap[primitives.AccountId](keyBalances, keyHolds, hashing.Blake128, sc.DecodeU128),
		// Freezes:          support.NewHashStorageMap[primitives.AccountId](keyBalances, keyFreezes, hashing.Blake128, sc.DecodeU128),
	}
}
