package balances

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
)

var (
	validGcJson             = "{\"balances\":{\"balances\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1]]}}"
	accountId, _            = types.NewAccountId(sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)...)
	balanceOne              = sc.NewU128(uint64(1))
	balanceOverMaxUint64, _ = sc.NewU128FromString("184467440737095516150")
)

func Test_GenesisConfig_CreateDefaultConfig(t *testing.T) {
	target := setupModule()

	expectedGc := []byte("{\"balances\":{\"balances\":[]}}")

	gc, err := target.CreateDefaultConfig()

	assert.NoError(t, err)
	assert.Equal(t, expectedGc, gc)
}

func Test_GenesisConfig_BuildConfig(t *testing.T) {
	for _, tt := range []struct {
		name               string
		gcJson             string
		expectedErr        error
		shouldAssertCalled bool
		tryMutateExistsErr error
		balance            sc.U128
	}{
		{
			name:               "valid",
			gcJson:             validGcJson,
			balance:            balanceOne,
			shouldAssertCalled: true,
		},
		{
			name:        "invalid genesis address",
			gcJson:      "{\"balances\":{\"balances\":[[1,1]]}}",
			expectedErr: errInvalidAddrValue,
		},
		{
			name:        "duplicate genesis balance",
			gcJson:      "{\"balances\":{\"balances\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1],[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1]]}}",
			expectedErr: errDuplicateBalancesInGenesis,
		},
		{
			name:        "invalid ss58 address",
			gcJson:      "{\"balances\":{\"balances\":[[\"invalid\",1]]}}",
			expectedErr: errors.New("expected at least 2 bytes in base58 decoded address"),
		},
		{
			name:        "invalid genesis balance",
			gcJson:      "{\"balances\":{\"balances\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",\"invalid\"]]}}",
			expectedErr: errInvalidBalanceValue,
		},
		{
			name:               "balance greater than MaxUint64",
			gcJson:             "{\"balances\":{\"balances\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",184467440737095516150]]}}",
			balance:            balanceOverMaxUint64,
			shouldAssertCalled: true,
		},
		{
			name:   "zero balances",
			gcJson: "{\"aura\":{\"authorities\":[]}}",
		},
		{
			name:        "balance below existential deposit",
			gcJson:      "{\"balances\":{\"balances\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",0]]}}",
			expectedErr: errBalanceBelowExistentialDeposit,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			target := setupModule()

			data := types.AccountData{
				Free:     tt.balance,
				Reserved: sc.NewU128(0),
				Frozen:   sc.NewU128(0),
				Flags:    types.DefaultExtraFlags,
			}

			mockStoredMap.On("IncProviders", accountId).Return(primitives.IncRefStatus(0), nil)
			mockStoredMap.On("Insert", accountId, data).Return(sc.Empty{}, nil)
			mockTotalIssuance.On("Put", tt.balance).Return()

			err := target.BuildConfig([]byte(tt.gcJson))

			assert.Equal(t, tt.expectedErr, err)
			if tt.shouldAssertCalled {
				mockStoredMap.AssertCalled(t, "IncProviders", accountId)
				mockStoredMap.AssertCalled(t, "Insert", accountId, data)
				mockTotalIssuance.AssertCalled(t, "Put", tt.balance)
			}
		})
	}
}
