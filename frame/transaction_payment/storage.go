package transaction_payment

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
)

var (
	keyTransactionPayment = []byte("TransactionPayment")
	keyNextFeeMultiplier  = []byte("NextFeeMultiplier")
)

var defaultMultiplierValue = sc.NewU128(1)

type storage struct {
	NextFeeMultiplier support.StorageValue[sc.U128]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		NextFeeMultiplier: support.NewHashStorageValueWithDefault(
			s,
			keyTransactionPayment,
			keyNextFeeMultiplier,
			sc.DecodeU128,
			&defaultMultiplierValue,
		),
	}
}
