package sudo

import (
	"github.com/LimeChain/gosemble/frame/support"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keySudo = []byte("Sudo")
	keyKey  = []byte("Key")
)

type storage struct {
	Key support.StorageValue[primitives.AccountId]
}

func newStorage() *storage {
	return &storage{
		Key: support.NewHashStorageValue(keySudo, keyKey, primitives.DecodeAccountId),
	}
}
