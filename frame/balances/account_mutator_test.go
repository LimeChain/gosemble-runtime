package balances

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_AccountMutator_ensureUpgraded(t *testing.T) {
	tests := []struct {
		name        string
		who         primitives.AccountId
		setupMocks  func()
		expectedRes bool
		expectedErr error
	}{
		{
			name: "error account store get",
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{}, expectedErr)
			},
			expectedErr: primitives.NewDispatchErrorOther(sc.Str(expectedErr.Error())),
		},
		{
			name: "is new logic",
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.DefaultAccountData()}, nil)
			},
			expectedRes: false,
		},
		{
			name: "error account store inc providers",
			setupMocks: func() {
				data := primitives.AccountData{Reserved: sc.NewU128(10)}
				acc := primitives.AccountInfo{Providers: 0, Data: data}
				mockStoredMap.On("Get", accId).Return(acc, nil)
				mockStoredMap.On("IncProviders", accId).Return(primitives.IncRefStatusCreated, expectedErr)
			},
			expectedErr: expectedErr,
		},
		{
			name: "error account store inc consumers without limit",
			setupMocks: func() {
				data := primitives.AccountData{Reserved: sc.NewU128(10)}
				acc := primitives.AccountInfo{Providers: 1, Data: data}
				mockStoredMap.On("Get", accId).Return(acc, nil)
				mockStoredMap.On("IncConsumersWithoutLimit", accId).Return(expectedErr)
			},
			expectedErr: expectedErr,
		},
		{
			name: "error account store try mutate exists",
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(accountInfo, nil)
				mockStoredMap.On("TryMutateExistsNoClosure", accId, mock.Anything).Return(expectedErr)
			},
			expectedErr: expectedErr,
		},
		{
			name: "happy path",
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{}, nil)
				mockStoredMap.On("TryMutateExistsNoClosure", accId, mock.Anything).Return(nil)
				mockStoredMap.On("DepositEvent", newEventUpgraded(moduleId, accId))
			},
			expectedRes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupModule()

			tt.setupMocks()

			res, err := target.ensureUpgraded(accId)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}
