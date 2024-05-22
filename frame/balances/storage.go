package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
)

var (
	keyBalances      = []byte("Balances")
	keyTotalIssuance = []byte("TotalIssuance")
)

type storage struct {
	TotalIssuance support.StorageValue[sc.U128]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		TotalIssuance: support.NewHashStorageValue(s, keyBalances, keyTotalIssuance, sc.DecodeU128),
	}
}
