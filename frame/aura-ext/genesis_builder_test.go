package aura_ext

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func Test_GenesisConfig_DefaultConfig(t *testing.T) {
	target := setupModule()

	result, err := target.CreateDefaultConfig()
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func Test_GenesisConfig_BuildConfig(t *testing.T) {
	target := setupModule()

	mockAuraModule.On("StorageAuthorities").Return(authorities, nil)
	mockAuthorities.On("Put", authorities).Return()

	err := target.BuildConfig([]byte{1})
	assert.Nil(t, err)

	mockAuraModule.AssertCalled(t, "StorageAuthorities")
	mockAuthorities.AssertCalled(t, "Put", authorities)
}

func Test_GenesisConfig_BuildConfig_Err(t *testing.T) {
	target := setupModule()
	expectErr := errors.New("err")

	mockAuraModule.On("StorageAuthorities").Return(authorities, expectErr)

	err := target.BuildConfig([]byte{1})
	assert.Equal(t, expectErr, err)

	mockAuraModule.AssertCalled(t, "StorageAuthorities")
}

func Test_GenesisConfig_BuildConfig_Empty(t *testing.T) {
	target := setupModule()

	mockAuraModule.On("StorageAuthorities").Return(sc.Sequence[primitives.Sr25519PublicKey]{}, nil)

	err := target.BuildConfig([]byte{1})
	assert.Equal(t, errAuraAuthoritiesEmpty, err)

	mockAuraModule.AssertCalled(t, "StorageAuthorities")
}
