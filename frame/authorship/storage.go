package authorship

import (
	"github.com/LimeChain/gosemble/frame/support"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyAuthorship = []byte("Authorship")
	keyAuthor     = []byte("Author")
)

type storage struct {
	Author support.StorageValue[primitives.AccountId]
}

func newStorage() *storage {
	return &storage{
		Author: support.NewHashStorageValue[primitives.AccountId](keyAuthorship, keyAuthor, primitives.DecodeAccountId),
	}
}
