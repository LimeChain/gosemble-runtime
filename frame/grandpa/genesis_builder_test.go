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
	validGcJson = "{\"grandpa\":{\"authorities\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1]]}}"
	accId, _    = types.NewAccountId(sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)...)
	authorities = sc.Sequence[types.Authority]{{Id: accId, Weight: sc.U64(1)}}
)

func Test_GenesisConfig_BuildConfig(t *testing.T) {
	for _, tt := range []struct {
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
			storageAuthorities: authorities,
			expectedErr:        errAuthoritiesAlreadyInitialized,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			mockStorageAuthorities.On("Get").Return(tt.storageAuthorities, tt.storageAuthoritiesGetErr)
			mockStorageAuthorities.On("Put", authorities).Return()

			err := target.BuildConfig([]byte(tt.gcJson))
			assert.Equal(t, tt.expectedErr, err)

			if tt.shouldAssertCalled {
				mockStorageAuthorities.AssertCalled(t, "Get")
				mockStorageAuthorities.AssertCalled(t, "Put", authorities)
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
