package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
)

var (
	keyBabe            = []byte("Babe")
	keyAuthorities     = []byte("Authorities")
	keyCurrentSlot     = []byte("CurrentSlot")
	keyEpochConfig     = []byte("EpochConfig")
	keyEpochIndex      = []byte("EpochIndex")
	keyGenesisSlot     = []byte("GenesisSlot")
	keyNextAuthorities = []byte("NextAuthorities")
	keyNextRandomness  = []byte("NextRandomness")
	keyRandomness      = []byte("Randomness")
	keySegmentIndex    = []byte("SegmentIndex")
)

func decodeAuthorities(buffer *bytes.Buffer) (sc.Sequence[Authority], error) {
	return sc.DecodeSequenceWith(buffer, DecodeAuthority)
}

func decodeRandomness(buffer *bytes.Buffer) (sc.FixedSequence[sc.U8], error) {
	return sc.DecodeFixedSequence[sc.U8](RandomnessLength, buffer)
}

type storage struct {
	Authorities     support.StorageValue[sc.Sequence[Authority]]
	EpochConfig     support.StorageValue[BabeEpochConfiguration]
	EpochIndex      support.StorageValue[sc.U64]
	GenesisSlot     support.StorageValue[Slot]
	CurrentSlot     support.StorageValue[Slot]
	NextAuthorities support.StorageValue[sc.Sequence[Authority]]
	NextEpochConfig support.StorageValue[BabeEpochConfiguration]
	NextRandomness  support.StorageValue[Randomness]
	Randomness      support.StorageValue[Randomness]
	SegmentIndex    support.StorageValue[sc.U32]
}

func newStorage() *storage {
	return &storage{
		Authorities:     support.NewHashStorageValue(keyBabe, keyAuthorities, decodeAuthorities),
		EpochConfig:     support.NewHashStorageValue(keyBabe, keyEpochConfig, DecodeBabeEpochConfiguration),
		EpochIndex:      support.NewHashStorageValue(keyBabe, keyEpochIndex, sc.DecodeU64),
		GenesisSlot:     support.NewHashStorageValue(keyBabe, keyGenesisSlot, sc.DecodeU64),
		CurrentSlot:     support.NewHashStorageValue(keyBabe, keyCurrentSlot, sc.DecodeU64),
		NextAuthorities: support.NewHashStorageValue(keyBabe, keyNextAuthorities, decodeAuthorities),
		NextEpochConfig: support.NewHashStorageValue(keyBabe, keyEpochConfig, DecodeBabeEpochConfiguration),
		NextRandomness:  support.NewHashStorageValue(keyBabe, keyNextRandomness, decodeRandomness),
		Randomness:      support.NewHashStorageValue(keyBabe, keyRandomness, decodeRandomness),
		SegmentIndex:    support.NewHashStorageValue(keyBabe, keySegmentIndex, sc.DecodeU32),
	}
}
