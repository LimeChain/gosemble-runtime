package babe

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/babe"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	epochDuration = sc.U64(1000)
	slotDuration  = epochDuration / 100

	genesisEpochConfig = babetypes.EpochConfiguration{
		C:            types.RationalValue{Numerator: 1, Denominator: 4},
		AllowedSlots: babetypes.NewPrimaryAndSecondaryPlainSlots(),
	}

	authorities = sc.Sequence[babetypes.Authority]{
		babetypes.Authority{
			Key:    types.Sr25519PublicKey{FixedSequence: sc.BytesToFixedSequenceU8([]byte{1, 2, 3})},
			Weight: sc.U64(1),
		},
	}

	randomness = babetypes.Randomness(sc.BytesToFixedSequenceU8([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))

	config = babetypes.Configuration{
		SlotDuration: slotDuration,
		EpochLength:  epochDuration,
		C:            genesisEpochConfig.C,
		Authorities:  authorities,
		Randomness:   randomness,
		AllowedSlots: genesisEpochConfig.AllowedSlots,
	}

	epoch = babetypes.Epoch{
		EpochIndex:  sc.U64(1),
		StartSlot:   babetypes.Slot(1),
		Duration:    epochDuration,
		Authorities: authorities,
		Randomness:  randomness,
		Config:      genesisEpochConfig,
	}

	expectedErr = errors.New("panic")
)

var target Module

var (
	mockBabe        *mocks.BabeModule
	mockMemoryUtils *mocks.MemoryTranslator
)

func setup() {
	mockBabe = new(mocks.BabeModule)
	mockMemoryUtils = new(mocks.MemoryTranslator)

	target = New(mockBabe, log.NewLogger())
	target.memUtils = mockMemoryUtils
}

func Test_Name(t *testing.T) {
	setup()

	assert.Equal(t, "BabeApi", target.Name())
}

func Test_Item(t *testing.T) {
	setup()

	hash := hashing.MustBlake2b8([]byte("BabeApi"))

	expected := types.ApiItem{
		Name:    sc.BytesToFixedSequenceU8(hash[:]),
		Version: 2,
	}

	assert.Equal(t, expected, target.Item())
}

func Test_Configuration_Empty_Config(t *testing.T) {
	setup()

	mockBabe.On("StorageEpochConfig").Return(babetypes.EpochConfiguration{}, nil)
	mockBabe.On("EpochConfig").Return(genesisEpochConfig)
	mockBabe.On("StorageAuthorities").Return(authorities, nil)
	mockBabe.On("StorageRandomness").Return(randomness, nil)
	mockBabe.On("SlotDuration").Return(slotDuration)
	mockBabe.On("EpochDuration").Return(epochDuration)
	mockMemoryUtils.On("BytesToOffsetAndSize", config.Bytes()).Return(int64(0))

	target.Configuration()

	mockBabe.AssertCalled(t, "StorageEpochConfig")
	mockBabe.AssertCalled(t, "EpochConfig")
	mockBabe.AssertCalled(t, "StorageAuthorities")
	mockBabe.AssertCalled(t, "StorageRandomness")
	mockBabe.AssertCalled(t, "SlotDuration")
	mockBabe.AssertCalled(t, "EpochDuration")
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", config.Bytes())
}

func Test_Configuration_Stored_Config(t *testing.T) {
	setup()

	someStoredConfig := babetypes.EpochConfiguration{C: types.RationalValue{Numerator: 1, Denominator: 1}}

	expectedConfig := config
	expectedConfig.C = someStoredConfig.C
	expectedConfig.AllowedSlots = someStoredConfig.AllowedSlots

	mockBabe.On("StorageEpochConfig").Return(someStoredConfig, nil)
	mockBabe.On("StorageAuthorities").Return(authorities, nil)
	mockBabe.On("StorageRandomness").Return(randomness, nil)
	mockBabe.On("SlotDuration").Return(slotDuration)
	mockBabe.On("EpochDuration").Return(epochDuration)
	mockMemoryUtils.On("BytesToOffsetAndSize", expectedConfig.Bytes()).Return(int64(0))

	target.Configuration()

	mockBabe.AssertCalled(t, "StorageEpochConfig")
	mockBabe.AssertNotCalled(t, "EpochConfig")
	mockBabe.AssertCalled(t, "StorageAuthorities")
	mockBabe.AssertCalled(t, "StorageRandomness")
	mockBabe.AssertCalled(t, "SlotDuration")
	mockBabe.AssertCalled(t, "EpochDuration")
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", expectedConfig.Bytes())
}

func Test_CurrentEpochStart(t *testing.T) {
	setup()

	slot := 5 * slotDuration

	mockBabe.On("CurrentEpochStart").Return(slot, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", slot.Bytes()).Return(int64(0))

	target.CurrentEpochStart()

	mockBabe.AssertCalled(t, "CurrentEpochStart")
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", slot.Bytes())
}

func Test_CurrentEpochStart_Panics(t *testing.T) {
	setup()

	mockBabe.On("CurrentEpochStart").Return(sc.U64(0), expectedErr)

	assert.PanicsWithValue(t, expectedErr.Error(), func() { target.CurrentEpochStart() })

	mockBabe.AssertCalled(t, "CurrentEpochStart")
	mockMemoryUtils.AssertNotCalled(t, "BytesToOffsetAndSize")
}

func Test_CurrentEpoch(t *testing.T) {
	setup()

	mockBabe.On("CurrentEpoch").Return(epoch, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", epoch.Bytes()).Return(int64(0))

	target.CurrentEpoch()

	mockBabe.AssertCalled(t, "CurrentEpoch")
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", epoch.Bytes())
}

func Test_CurrentEpoch_Panics(t *testing.T) {
	setup()

	mockBabe.On("CurrentEpoch").Return(babe.Epoch{}, expectedErr)

	assert.PanicsWithValue(t, expectedErr.Error(), func() { target.CurrentEpoch() })

	mockBabe.AssertCalled(t, "CurrentEpoch")
	mockMemoryUtils.AssertNotCalled(t, "BytesToOffsetAndSize", epoch.Bytes())
}

func Test_NextEpoch(t *testing.T) {
	setup()

	mockBabe.On("NextEpoch").Return(epoch, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", epoch.Bytes()).Return(int64(0))

	target.NextEpoch()

	mockBabe.AssertCalled(t, "NextEpoch")
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", epoch.Bytes())
}

func Test_NextEpoch_Panics(t *testing.T) {
	setup()

	mockBabe.On("NextEpoch").Return(babe.Epoch{}, expectedErr)

	assert.PanicsWithValue(t, expectedErr.Error(), func() { target.NextEpoch() })

	mockBabe.AssertCalled(t, "NextEpoch")
	mockMemoryUtils.AssertNotCalled(t, "BytesToOffsetAndSize", epoch.Bytes())
}

func Test_Module_Metadata(t *testing.T) {
	setup()

	expect := types.RuntimeApiMetadata{
		Name: ApiModuleName,
		Methods: sc.Sequence[primitives.RuntimeApiMethodMetadata]{
			primitives.RuntimeApiMethodMetadata{
				Name:   "configuration",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
				Output: sc.ToCompact(metadata.TypesBabeConfiguration),
				Docs:   sc.Sequence[sc.Str]{""},
			},
			primitives.RuntimeApiMethodMetadata{
				Name:   "current_epoch_start",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
				Output: sc.ToCompact(metadata.PrimitiveTypesU64),
				Docs:   sc.Sequence[sc.Str]{""},
			},
			primitives.RuntimeApiMethodMetadata{
				Name:   "current_epoch",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
				Output: sc.ToCompact(metadata.TypesBabeEpoch),
				Docs:   sc.Sequence[sc.Str]{""},
			},
			primitives.RuntimeApiMethodMetadata{
				Name:   "next_epoch",
				Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
				Output: sc.ToCompact(metadata.TypesBabeEpoch),
				Docs:   sc.Sequence[sc.Str]{""},
			},
		},
		Docs: sc.Sequence[sc.Str]{"Babe consensus API module."},
	}

	assert.Equal(t, expect, target.Metadata())
}
