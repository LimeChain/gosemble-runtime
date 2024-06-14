package aura_ext

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyAura        = []byte("AuraExt")
	keyAuthorities = []byte("Authorities")
	keyCurrentSlot = []byte("CurrentSlot")
)

type storage struct {
	Authorities support.StorageValue[sc.Sequence[types.Sr25519PublicKey]]
	SlotInfo    support.StorageValue[SlotInfo]
}

func newStorage(s io.Storage) *storage {
	return &storage{
		Authorities: support.NewHashStorageValue(s, keyAura, keyAuthorities, types.DecodeSequenceSr25519PublicKey),
		SlotInfo:    support.NewHashStorageValue(s, keyAura, keyCurrentSlot, DecodeSlotInfo),
	}
}
