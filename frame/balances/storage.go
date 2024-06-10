package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
)

var (
	keyBalances         = []byte("Balances")
	keyInactiveIssuance = []byte("InactiveIssuance")
	keyTotalIssuance    = []byte("TotalIssuance")
)

type storage struct {
	InactiveIssuance support.StorageValue[sc.U128]
	TotalIssuance    support.StorageValue[sc.U128]
}

func newStorage() *storage {
	return &storage{
		InactiveIssuance: support.NewHashStorageValue(keyBalances, keyInactiveIssuance, sc.DecodeU128),
		TotalIssuance:    support.NewHashStorageValue(keyBalances, keyTotalIssuance, sc.DecodeU128),
	}
}
