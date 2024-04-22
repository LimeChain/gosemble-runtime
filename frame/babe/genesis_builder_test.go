package babe

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	gcJsonInvalid         = "{'}"
	gcJsonInvalidAddress  = "{\"babe\":{\"authorities\":[\"abc\"],\"epochConfig\":{\"c\":[0,0],\"allowed_slots\":\"\"}}}"
	gcJsonInvalidConfig   = "{\"babe\":{\"authorities\":[],\"epochConfig\":{\"c\":[0,0],\"allowed_slots\":\"xyz\"}}}"
	gcJsonDefault         = "{\"babe\":{\"authorities\":[],\"epochConfig\":{\"c\":[0,0],\"allowed_slots\":\"\"}}}"
	gcJsonNoAuthorities   = "{\"babe\":{\"authorities\":[],\"epochConfig\":{\"c\":[2,3],\"allowed_slots\":\"PrimaryAndSecondaryVRFSlots\"}}}"
	gcJsonSomeAuthorities = "{\"babe\":{\"authorities\":[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"],\"epochConfig\":{\"c\":[1,4],\"allowed_slots\":\"PrimarySlots\"}}}"
)

func Test_Babe_CreateDefaultConfig(t *testing.T) {
	target := setupModule()

	gcJson, err := target.CreateDefaultConfig()

	assert.NoError(t, err)

	assert.Equal(t, gcJsonDefault, string(gcJson))
}

func Test_Babe_BuildConfig_Invalid_Json(t *testing.T) {
	target := setupModule()

	err := target.BuildConfig([]byte(gcJsonInvalid))

	assert.Error(t, err)

	mockStorageSegmentIndex.AssertNotCalled(t, "Put", sc.U32(0))
	mockStorageAuthorities.AssertNotCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageEpochConfig.AssertNotCalled(t, "Put", epochConfig)
}

func Test_Babe_BuildConfig_Invalid_Address(t *testing.T) {
	target := setupModule()

	err := target.BuildConfig([]byte(gcJsonInvalidAddress))

	assert.Equal(t, errors.New("checksum mismatch: expected [126 165] but got [185 123]"), err)

	mockStorageSegmentIndex.AssertNotCalled(t, "Put", sc.U32(0))
	mockStorageAuthorities.AssertNotCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageEpochConfig.AssertNotCalled(t, "Put", epochConfig)
}

func Test_Babe_BuildConfig_Invalid_Config(t *testing.T) {
	target := setupModule()

	err := target.BuildConfig([]byte(gcJsonInvalidConfig))

	assert.Equal(t, errors.New("invalid 'AllowedSlots' type"), err)

	mockStorageSegmentIndex.AssertNotCalled(t, "Put", sc.U32(0))
	mockStorageAuthorities.AssertNotCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageEpochConfig.AssertNotCalled(t, "Put", epochConfig)
}

func Test_Babe_BuildConfig_No_Authorities(t *testing.T) {
	epochConfig := babetypes.EpochConfiguration{
		C:            primitives.RationalValue{Numerator: sc.U64(2), Denominator: sc.U64(3)},
		AllowedSlots: babetypes.NewPrimaryAndSecondaryVRFSlots(),
	}

	target := setupModule()

	mockStorageSegmentIndex.On("Put", sc.U32(0)).Return()
	mockStorageEpochConfig.On("Put", epochConfig).Return()

	err := target.BuildConfig([]byte(gcJsonNoAuthorities))

	assert.NoError(t, err)
	mockStorageSegmentIndex.AssertCalled(t, "Put", sc.U32(0))
	mockStorageAuthorities.AssertNotCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageEpochConfig.AssertCalled(t, "Put", epochConfig)
}

func Test_Babe_BuildConfig_Some_Authorities(t *testing.T) {
	target := setupModule()

	mockStorageSegmentIndex.On("Put", sc.U32(0)).Return()
	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](nil), nil)
	mockStorageAuthorities.On("Put", authorities).Return()
	mockStorageNextAuthorities.On("Put", authorities).Return()
	mockStorageEpochConfig.On("Put", epochConfig).Return()

	err := target.BuildConfig([]byte(gcJsonSomeAuthorities))

	assert.NoError(t, err)

	mockStorageSegmentIndex.AssertCalled(t, "Put", sc.U32(0))
	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertCalled(t, "Put", authorities)
	mockStorageEpochConfig.AssertCalled(t, "Put", epochConfig)
}

func Test_Babe_BuildConfig_Authorities_Error(t *testing.T) {
	target := setupModule()

	someError := errors.New("some error")

	mockStorageSegmentIndex.On("Put", sc.U32(0)).Return()
	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](nil), someError)

	err := target.BuildConfig([]byte(gcJsonSomeAuthorities))

	assert.Error(t, someError, err)

	mockStorageSegmentIndex.AssertCalled(t, "Put", sc.U32(0))
	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
	mockStorageAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageNextAuthorities.AssertNotCalled(t, "Put", authorities)
	mockStorageEpochConfig.AssertNotCalled(t, "Put", epochConfig)
}
