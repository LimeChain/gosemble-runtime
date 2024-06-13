package sudo

import (
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keySudo = []byte("Sudo")
	keyKey  = []byte("Key")
)

type storage struct {
	Key support.StorageValue[primitives.AccountId]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		Key: support.NewHashStorageValue(s, keySudo, keyKey, primitives.DecodeAccountId),
	}
}
