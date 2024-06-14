package timestamp

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
)

var (
	keyTimestamp = []byte("Timestamp")
	keyDidUpdate = []byte("DidUpdate")
	keyNow       = []byte("Now")
)

type storage struct {
	Now       support.StorageValue[sc.U64]
	DidUpdate support.StorageValue[sc.Bool]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		Now:       support.NewHashStorageValue(s, keyTimestamp, keyNow, sc.DecodeU64),
		DidUpdate: support.NewHashStorageValue(s, keyTimestamp, keyDidUpdate, sc.DecodeBool),
	}
}
