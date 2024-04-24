package grandpa

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
)

var (
	validGcJson   = "{\"grandpa\":{\"authorities\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1]]}}"
	accountId, _  = types.NewAccountId(sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)...)
	authorityList = sc.Sequence[types.Authority]{{Id: accountId, Weight: 1}}
)

func Test_GenesisConfig_BuildConfig(t *testing.T) {
	for _, testExample := range []struct {
		name                     string
		gcJson                   string
		expectedErr              error
		shouldAssertCalled       bool
		storageAuthorities       sc.Sequence[types.Authority]
		storageAuthoritiesGetErr error
	}{
		{
			name:               "valid",
			gcJson:             validGcJson,
			shouldAssertCalled: true,
		},
		{
			name:               "duplicate genesis address",
			gcJson:             "{\"grandpa\":{\"authorities\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1],[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1]]}}",
			shouldAssertCalled: true,
		},
		{
			name:        "invalid genesis address",
			gcJson:      "{\"grandpa\":{\"authorities\":[[1,1]]}}",
			expectedErr: errInvalidAddrValue,
		},
		{
			name:        "invalid ss58 address",
			gcJson:      "{\"grandpa\":{\"authorities\":[[\"invalid\",1]]}}",
			expectedErr: errors.New("expected at least 2 bytes in base58 decoded address"),
		},
		{
			name:        "invalid genesis weight",
			gcJson:      "{\"grandpa\":{\"authorities\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",\"invalid\"]]}}",
			expectedErr: errInvalidWeightValue,
		},
		{
			name:   "zero authorities",
			gcJson: "{\"grandpa\":{\"authorities\":[]}}",
		},
		{
			name:                     "storage authorities error on get",
			gcJson:                   validGcJson,
			storageAuthoritiesGetErr: errors.New("err"),
			expectedErr:              errors.New("err"),
		},
		{
			name:               "storage authorities already initialized",
			gcJson:             validGcJson,
			storageAuthorities: authorityList,
			expectedErr:        errAuthoritiesAlreadyInitialized,
		},
	} {
		t.Run(testExample.name, func(t *testing.T) {
			setup()

			mockStorageCurrentSetId.On("Put", sc.U64(0)).Return(nil)
			mockStorageAuthorities.On("Get").Return(testExample.storageAuthorities, testExample.storageAuthoritiesGetErr)
			mockStorageAuthorities.On("Put", authorityList).Return(nil)
			mockStorageSetIdSession.On("Put", sc.U64(0), sc.U32(0)).Return(nil)

			err := target.BuildConfig([]byte(testExample.gcJson))

			assert.Equal(t, testExample.expectedErr, err)

			if testExample.shouldAssertCalled {
				mockStorageCurrentSetId.AssertCalled(t, "Put", sc.U64(0))
				mockStorageAuthorities.AssertCalled(t, "Get")
				mockStorageAuthorities.AssertCalled(t, "Put", authorityList)
				mockStorageSetIdSession.AssertCalled(t, "Put", sc.U64(0), sc.U32(0))
			}
		})
	}
}

func Test_CreateDefaultConfig(t *testing.T) {
	setup()

	expectedGc := []byte("{\"grandpa\":{\"authorities\":[]}}")

	gc, err := target.CreateDefaultConfig()
	assert.NoError(t, err)
	assert.Equal(t, expectedGc, gc)
}
