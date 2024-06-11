package sudo

import (
	"errors"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	bytesAccount    = []byte{0xd4, 0x35, 0x93, 0xc7, 0x15, 0xfd, 0xd3, 0x1c, 0x61, 0x14, 0x1a, 0xbd, 0x4, 0xa9, 0x9f, 0xd6, 0x82, 0x2c, 0x85, 0x58, 0x85, 0x4c, 0xcd, 0xe3, 0x9a, 0x56, 0x84, 0xe7, 0xa5, 0x6d, 0xa2, 0x7d}
	fixedSeqAccount = sc.BytesToFixedSequenceU8(bytesAccount)
	accountId, _    = primitives.NewAccountId(fixedSeqAccount...)
)

func Test_GenesisConfig_CreateDefaultConfig(t *testing.T) {
	target := setupModule()
	expectedGc := []byte("{\"sudo\":{\"key\":\"\"}}")

	gc, err := target.CreateDefaultConfig()
	assert.NoError(t, err)
	assert.Equal(t, expectedGc, gc)
}

func Test_GenesisConfig_BuildConfig(t *testing.T) {
	gcJson := "{\"sudo\":{\"key\":\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"}}"
	target := setupModule()

	mockStorageKey.On("Put", accountId)

	err := target.BuildConfig([]byte(gcJson))
	assert.Nil(t, err)

	mockStorageKey.AssertCalled(t, "Put", accountId)
}

func Test_GenesisConfig_BuildConfig_Empty(t *testing.T) {
	target := setupModule()

	err := target.BuildConfig([]byte("{}"))
	assert.Equal(t, errors.New("expected at least 2 bytes in base58 decoded address"), err)
}
