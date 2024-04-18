package aura_ext

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
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

func newStorage() *storage {
	return &storage{
		Authorities: support.NewHashStorageValue(keyAura, keyAuthorities, types.DecodeSequenceSr25519PublicKey),
		SlotInfo:    support.NewHashStorageValue(keyAura, keyCurrentSlot, DecodeSlotInfo),
	}
}
