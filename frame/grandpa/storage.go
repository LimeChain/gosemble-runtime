package grandpa

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyGrandpa       = []byte("Grandpa")
	keyAuthorities   = []byte("Authorities") // TODO: Gossamer uses ":grandpa_authorities"
	keyCurrentSetId  = []byte("CurrentSetId")
	keyStalled       = []byte("Stalled")
	keyPendingChange = []byte("PendingChange")
	keyState         = []byte("State")
	keySetIdSession  = []byte("SetIdSession")
	keyNextForced    = []byte("NextForced")
)

type storage struct {
	Authorities   support.StorageValue[sc.Sequence[primitives.Authority]]
	CurrentSetId  support.StorageValue[sc.U64]
	Stalled       support.StorageValue[primitives.Tuple2U64]
	PendingChange support.StorageValue[StoredPendingChange]
	State         support.StorageValue[StoredState]
	SetIdSession  support.StorageMap[sc.U64, sc.U32]
	NextForced    support.StorageValue[sc.U64]
}

func newStorage() *storage {
	hashing := io.NewHashing()

	return &storage{
		// Authorities:   support.NewSimpleStorageValue(keyAuthorities, primitives.DecodeAuthorityList),
		Authorities:   support.NewHashStorageValue(keyGrandpa, keyAuthorities, primitives.DecodeAuthorityList),
		CurrentSetId:  support.NewHashStorageValue(keyGrandpa, keyCurrentSetId, sc.DecodeU64),
		Stalled:       support.NewHashStorageValue(keyGrandpa, keyStalled, primitives.DecodeTuple2U64),
		PendingChange: support.NewHashStorageValue(keyGrandpa, keyPendingChange, DecodeStoredPendingChange),
		State:         support.NewHashStorageValueWithDefault(keyGrandpa, keyState, DecodeStoredState, defaultState()),
		SetIdSession:  support.NewHashStorageMap[sc.U64, sc.U32](keyGrandpa, keySetIdSession, hashing.Twox64, sc.DecodeU32),
		NextForced:    support.NewHashStorageValue(keyGrandpa, keyNextForced, sc.DecodeU64),
	}
}

func defaultState() *StoredState {
	s := NewStoredStateLive()
	return &s
}
