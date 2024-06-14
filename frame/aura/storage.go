package aura

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyAura        = []byte("Aura")
	keyAuthorities = []byte("Authorities")
	keyCurrentSlot = []byte("CurrentSlot")
)

type storage struct {
	Authorities support.StorageValue[sc.Sequence[types.Sr25519PublicKey]]
	CurrentSlot support.StorageValue[sc.U64]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		Authorities: support.NewHashStorageValue(s, keyAura, keyAuthorities, types.DecodeSequenceSr25519PublicKey),
		CurrentSlot: support.NewHashStorageValue(s, keyAura, keyCurrentSlot, sc.DecodeU64),
	}
}
