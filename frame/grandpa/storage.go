package grandpa

import (
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyGrandpaAuthorities = []byte(":grandpa_authorities")
)

type storage struct {
	Authorities support.StorageValue[primitives.VersionedAuthorityList]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		Authorities: support.NewSimpleStorageValue(s, keyGrandpaAuthorities, primitives.DecodeVersionedAuthorityList),
	}
}
