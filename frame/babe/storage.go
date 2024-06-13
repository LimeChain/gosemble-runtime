package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/io"

	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keyBabe                     = []byte("Babe")
	keyAuthorities              = []byte("Authorities")
	keyAuthorVrfRandomness      = []byte("AuthorVrfRandomness")
	keyCurrentSlot              = []byte("CurrentSlot")
	keyEpochConfig              = []byte("EpochConfig")
	keyEpochIndex               = []byte("EpochIndex")
	keyEpochStart               = []byte("EpochStart")
	keyGenesisSlot              = []byte("GenesisSlot")
	keyInitialized              = []byte("Initialized")
	keyLateness                 = []byte("Lateness")
	keyNextAuthorities          = []byte("NextAuthorities")
	keyNextEpochConfig          = []byte("NextEpochConfig")
	keyNextRandomness           = []byte("NextRandomness")
	keyPendingEpochConfigChange = []byte("PendingEpochConfigChange")
	keyRandomness               = []byte("Randomness")
	keySegmentIndex             = []byte("SegmentIndex")
	keySkippedEpochs            = []byte("SkippedEpochs")
	keyUnderConstruction        = []byte("UnderConstruction")
)

var defaultRandomnessValue = babetypes.NewRandomness()

type storage struct {
	Authorities              support.StorageValue[sc.Sequence[primitives.Authority]]
	AuthorVrfRandomness      support.StorageValue[sc.Option[babetypes.Randomness]]
	CurrentSlot              support.StorageValue[babetypes.Slot]
	EpochConfig              support.StorageValue[babetypes.EpochConfiguration]
	EpochIndex               support.StorageValue[sc.U64]
	EpochStart               support.StorageValue[babetypes.EpochStartBlocks]
	GenesisSlot              support.StorageValue[babetypes.Slot]
	Initialized              support.StorageValue[sc.Option[babetypes.PreDigest]]
	Lateness                 support.StorageValue[sc.U64]
	NextAuthorities          support.StorageValue[sc.Sequence[primitives.Authority]]
	NextEpochConfig          support.StorageValue[babetypes.EpochConfiguration]
	NextRandomness           support.StorageValue[babetypes.Randomness]
	PendingEpochConfigChange support.StorageValue[NextConfigDescriptor]
	Randomness               support.StorageValue[babetypes.Randomness]
	SegmentIndex             support.StorageValue[sc.U32]
	SkippedEpochs            support.StorageValue[sc.FixedSequence[babetypes.SkippedEpoch]]
	UnderConstruction        support.StorageMap[sc.U32, babetypes.Randomness]
}

func newStorage(s io.Storage) *storage {
	hashing := io.NewHashing()

	return &storage{
		Authorities:              support.NewHashStorageValue(s, keyBabe, keyAuthorities, primitives.DecodeAuthorityList),
		AuthorVrfRandomness:      support.NewHashStorageValue(s, keyBabe, keyAuthorVrfRandomness, decodeOptionRandomness),
		CurrentSlot:              support.NewHashStorageValue(s, keyBabe, keyCurrentSlot, sc.DecodeU64),
		EpochConfig:              support.NewHashStorageValue(s, keyBabe, keyEpochConfig, babetypes.DecodeEpochConfiguration),
		EpochIndex:               support.NewHashStorageValue(s, keyBabe, keyEpochIndex, sc.DecodeU64),
		EpochStart:               support.NewHashStorageValue(s, keyBabe, keyEpochStart, babetypes.DecodeEpochStartBlocks),
		GenesisSlot:              support.NewHashStorageValue(s, keyBabe, keyGenesisSlot, sc.DecodeU64),
		Initialized:              support.NewHashStorageValue(s, keyBabe, keyInitialized, decodePreDigest),
		Lateness:                 support.NewHashStorageValue(s, keyBabe, keyLateness, sc.DecodeU64),
		NextAuthorities:          support.NewHashStorageValue(s, keyBabe, keyNextAuthorities, primitives.DecodeAuthorityList),
		NextEpochConfig:          support.NewHashStorageValue(s, keyBabe, keyNextEpochConfig, babetypes.DecodeEpochConfiguration),
		NextRandomness:           support.NewHashStorageValueWithDefault(s, keyBabe, keyNextRandomness, decodeRandomness, &defaultRandomnessValue),
		PendingEpochConfigChange: support.NewHashStorageValue(s, keyBabe, keyPendingEpochConfigChange, DecodeNextConfigDescriptor),
		Randomness:               support.NewHashStorageValueWithDefault(s, keyBabe, keyRandomness, decodeRandomness, &defaultRandomnessValue),
		SegmentIndex:             support.NewHashStorageValue(s, keyBabe, keySegmentIndex, sc.DecodeU32),
		SkippedEpochs:            support.NewHashStorageValue(s, keyBabe, keySkippedEpochs, decodeSkippedEpochs),
		UnderConstruction:        support.NewHashStorageMap[sc.U32, babetypes.Randomness](s, keyBabe, keyUnderConstruction, hashing.Twox64, decodeRandomness),
	}
}

func decodeRandomness(buffer *bytes.Buffer) (sc.FixedSequence[sc.U8], error) {
	return sc.DecodeFixedSequence[sc.U8](babetypes.RandomnessLength, buffer)
}

func decodeOptionRandomness(buffer *bytes.Buffer) (sc.Option[sc.FixedSequence[sc.U8]], error) {
	return sc.DecodeOptionWith(buffer, decodeRandomness)
}

func decodePreDigest(buffer *bytes.Buffer) (sc.Option[babetypes.PreDigest], error) {
	return sc.DecodeOptionWith(buffer, babetypes.DecodePreDigest)
}

func decodeSkippedEpochs(buffer *bytes.Buffer) (sc.FixedSequence[babetypes.SkippedEpoch], error) {
	return sc.DecodeFixedSequence[babetypes.SkippedEpoch](SkippedEpochsBound, buffer)
}
