package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
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

func newStorage(s io.Storage) *storage {
	return &storage{
		InactiveIssuance: support.NewHashStorageValue(s, keyBalances, keyInactiveIssuance, sc.DecodeU128),
		TotalIssuance:    support.NewHashStorageValue(s, keyBalances, keyTotalIssuance, sc.DecodeU128),
	}
}
