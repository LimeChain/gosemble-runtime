package parachain_info

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
)

var (
	keyParachainInfo = []byte("ParachainInfo")
	keyParachainId   = []byte("ParachainId")
)

type storage struct {
	ParachainId support.StorageValue[sc.U32]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		ParachainId: support.NewHashStorageValueWithDefault(s, keyParachainInfo, keyParachainId, sc.DecodeU32, &defaultParachainId),
	}
}
