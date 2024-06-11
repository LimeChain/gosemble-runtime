package main

import (
	"bytes"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

var (
	pubKey1 = primitives.Sr25519PublicKey{FixedSequence: sc.NewFixedSequence[sc.U8](32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)}
	pubKey2 = primitives.Sr25519PublicKey{FixedSequence: sc.NewFixedSequence[sc.U8](32, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1)}

	slotDuration        = sc.U64(2000) // milliseconds
	epochLength         = sc.U64(200)  // slots
	probabilityConstant = primitives.Tuple2U64{First: 2, Second: 3}
	allowedSlots        = babetypes.NewPrimaryAndSecondaryVRFSlots()
	authorities         = sc.Sequence[primitives.Authority]{
		primitives.Authority{Id: primitives.AccountId(pubKey1), Weight: 1},
		primitives.Authority{Id: primitives.AccountId(pubKey2), Weight: 2},
	}
	randomness = babetypes.Randomness(sc.NewFixedSequence[sc.U8](32, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2))
)

var (
	epochConfig = babetypes.EpochConfiguration{
		C:            probabilityConstant,
		AllowedSlots: allowedSlots,
	}
)

func Test_Babe_Configuration_Genesis(t *testing.T) {
	rt, _ := testhelpers.NewRuntimeInstance(t)

	result, err := rt.Exec("BabeApi_configuration", []byte{})
	assert.NoError(t, err)

	config := gossamertypes.BabeConfiguration{}
	err = scale.Unmarshal(result, &config)
	assert.NoError(t, err)

	expectedConfig := gossamertypes.BabeConfiguration{
		SlotDuration:       uint64(2000),
		EpochLength:        uint64(200),
		C1:                 1,
		C2:                 4,
		GenesisAuthorities: []gossamertypes.AuthorityRaw(nil),
		Randomness:         [32]byte{},
		SecondarySlots:     uint8(0),
	}

	assert.Equal(t, expectedConfig, config)
}

func Test_Babe_Configuration(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyEpochConfigHash...), epochConfig.Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyAuthoritiesHash...), authorities.Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyRandomnessHash...), randomness.Bytes())

	result, err := rt.Exec("BabeApi_configuration", []byte{})
	assert.NoError(t, err)

	config := gossamertypes.BabeConfiguration{}
	err = scale.Unmarshal(result, &config)
	assert.NoError(t, err)

	expectedConfig := gossamertypes.BabeConfiguration{
		SlotDuration: uint64(2000),
		EpochLength:  uint64(200),
		C1:           2,
		C2:           3,
		GenesisAuthorities: []gossamertypes.AuthorityRaw{
			{
				Key:    [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				Weight: uint64(1),
			},
			{
				Key:    [32]uint8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				Weight: uint64(2),
			},
		},
		Randomness:     [32]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		SecondarySlots: uint8(2),
	}

	assert.Equal(t, expectedConfig, config)
}

func Test_Babe_CurrentEpochStart(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyEpochIndexHash...), sc.U64(2).Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyGenesisSlotHash...), sc.U64(20).Bytes())

	result, err := rt.Exec("BabeApi_current_epoch_start", []byte{})
	assert.NoError(t, err)

	epochStart := uint64(0)
	err = scale.Unmarshal(result, &epochStart)
	assert.NoError(t, err)

	assert.Equal(t, uint64(420), epochStart)
}

func Test_Babe_CurrentEpoch(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyEpochIndexHash...), sc.U64(2).Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyGenesisSlotHash...), sc.U64(20).Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyAuthoritiesHash...), authorities.Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyRandomnessHash...), randomness.Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyEpochConfigHash...), epochConfig.Bytes())

	result, err := rt.Exec("BabeApi_current_epoch", []byte{})
	assert.NoError(t, err)

	epoch, err := babetypes.DecodeEpoch(bytes.NewBuffer(result))
	assert.NoError(t, err)

	expectedEpoch := babetypes.Epoch{
		EpochIndex:  sc.U64(2),
		StartSlot:   sc.U64(420),
		Duration:    sc.U64(200),
		Authorities: authorities,
		Randomness:  randomness,
		Config:      epochConfig,
	}

	assert.Equal(t, expectedEpoch, epoch)
}

func Test_Babe_NextEpoch(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyEpochIndexHash...), sc.U64(2).Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyGenesisSlotHash...), sc.U64(20).Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyNextAuthoritiesHash...), authorities.Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyNextRandomnessHash...), randomness.Bytes())
	(*storage).Put(append(testhelpers.KeyBabeHash, testhelpers.KeyNextEpochConfigHash...), epochConfig.Bytes())

	result, err := rt.Exec("BabeApi_next_epoch", []byte{})
	assert.NoError(t, err)

	epoch, err := babetypes.DecodeEpoch(bytes.NewBuffer(result))
	assert.NoError(t, err)

	expectedEpoch := babetypes.Epoch{
		EpochIndex:  sc.U64(3),
		StartSlot:   sc.U64(620),
		Duration:    sc.U64(200),
		Authorities: authorities,
		Randomness:  randomness,
		Config:      epochConfig,
	}

	assert.Equal(t, expectedEpoch, epoch)
}

// TODO: once implemented
// BabeApi_generate_key_ownership_proof
// BabeApi_submit_report_equivocation_unsigned_extrinsic
