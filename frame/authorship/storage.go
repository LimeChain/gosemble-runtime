package authorship

import (
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyAuthorship = []byte("Authorship")
	keyAuthor     = []byte("Author")
)

type storage struct {
	Author support.StorageValue[primitives.AccountId]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		Author: support.NewHashStorageValue[primitives.AccountId](s, keyAuthorship, keyAuthor, primitives.DecodeAccountId),
	}
}
