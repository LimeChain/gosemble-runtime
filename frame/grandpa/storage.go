package grandpa

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyGrandpaAuthorities = []byte(":grandpa_authorities")
)

type storage struct {
	Authorities support.StorageValue[sc.Sequence[primitives.Authority]]
}

func newStorage() *storage {
	return &storage{
		Authorities: support.NewSimpleStorageValue(keyGrandpaAuthorities, primitives.DecodeAuthorityList),
	}
}
