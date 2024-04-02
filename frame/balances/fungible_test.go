package balances

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Fungible_CanWithdraw(t *testing.T) {
	tests := []struct {
		name        string
		amount      primitives.Balance
		expectedRes balancestypes.WithdrawConsequence
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "zero amount",
			amount:      constants.Zero,
			expectedRes: balancestypes.WithdrawConsequenceSuccess,
			setupMocks:  func() {},
		},
		{
			name:        "total Issuance Get error",
			amount:      sc.NewU128(100),
			expectedErr: expectedErr,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1), expectedErr)
			},
		},
		{
			name:        "total Issuance Underflow",
			amount:      sc.NewU128(100),
			expectedRes: balancestypes.WithdrawConsequenceUnderflow,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1), nil)
			},
		},
		{
			name:        "storedMap Get error",
			amount:      sc.NewU128(100),
			expectedErr: expectedErr,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{}, expectedErr)
			},
		},
		{
			name:        "insufficient balance",
			amount:      sc.NewU128(1000),
			expectedRes: balancestypes.WithdrawConsequenceBalanceLow,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(500)}}, nil)
			},
		},
		{
			name:        "error reducible balance",
			amount:      sc.NewU128(100),
			expectedRes: balancestypes.WithdrawConsequenceBalanceLow,
			// expectedErr: expectedErr,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(500)}}, nil).Once()
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{}, expectedErr).Once()
			},
		},
		{
			name:        "insufficient reducible balance",
			amount:      sc.NewU128(100),
			expectedRes: balancestypes.WithdrawConsequenceFrozen,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(500), Frozen: sc.NewU128(499)}}, nil)
			},
		},
		{
			name:        "lower than ED && CanDecProviders",
			amount:      sc.NewU128(500),
			expectedRes: balancestypes.WithdrawConsequenceReducedToZero,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Providers: 1, Data: primitives.AccountData{Free: sc.NewU128(500)}}, nil)
			},
		},
		{
			name:        "success",
			amount:      sc.NewU128(100),
			expectedRes: balancestypes.WithdrawConsequenceSuccess,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(500)}}, nil)
			},
		},
		{
			name:        "new Balance below frozen",
			amount:      sc.NewU128(100),
			expectedRes: balancestypes.WithdrawConsequenceFrozen,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(101), Frozen: sc.NewU128(500), Reserved: sc.NewU128(401)}}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupFungible()
			tt.setupMocks()

			got, err := target.canWithdraw(accId, tt.amount)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.Equal(t, tt.expectedRes, got)
			}
		})
	}
}

func Test_Fungible_CanDeposit(t *testing.T) {
	tests := []struct {
		name        string
		amount      primitives.Balance
		minted      bool
		expectedRes balancestypes.DepositConsequence
		expectedErr error
		setupMocks  func()
	}{
		{
			name:        "zero amount",
			amount:      constants.Zero,
			expectedRes: balancestypes.DepositConsequenceSuccess,
			setupMocks:  func() {},
		},
		{
			name:        "minted with successful deposit",
			amount:      sc.NewU128(100),
			minted:      true,
			expectedRes: balancestypes.DepositConsequenceSuccess,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(500)}}, nil)
			},
		},
		{
			name:        "minted with overflow",
			amount:      sc.MaxU128(),
			minted:      true,
			expectedRes: balancestypes.DepositConsequenceOverflow,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1000000), nil)
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(500)}}, nil)
			},
		},
		{
			name:        "existing account with successful deposit",
			amount:      sc.NewU128(100),
			expectedRes: balancestypes.DepositConsequenceSuccess,
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(500)}}, nil)
			},
		},
		{
			name:        "existing account with overflow",
			amount:      sc.MaxU128(),
			expectedRes: balancestypes.DepositConsequenceOverflow,
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(500)}}, nil)
			},
		},
		{
			name:        "existing account with below minimum deposit",
			amount:      sc.NewU128(2),
			expectedRes: balancestypes.DepositConsequenceBelowMinimum,
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(1)}}, nil)
			},
		},
		{
			name:        "total Issuance Get error",
			amount:      sc.NewU128(2),
			minted:      true,
			expectedErr: expectedErr,
			setupMocks: func() {
				mockTotalIssuance.On("Get").Return(sc.NewU128(1), expectedErr)
			},
		},
		{
			name:        "storedMap Get error",
			amount:      sc.NewU128(2),
			expectedErr: expectedErr,
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{}, expectedErr)
			},
		},
		{
			name:        "acc Reserved Overflow",
			amount:      sc.MaxU128(),
			expectedRes: balancestypes.DepositConsequenceOverflow,
			setupMocks: func() {
				mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Reserved: sc.MaxU128()}}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupFungible()
			target.constants.ExistentialDeposit = sc.NewU128(10)
			tt.setupMocks()

			got, err := target.canDeposit(accId, tt.amount, tt.minted)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.Equal(t, tt.expectedRes, got)
			}
		})
	}
}

func Test_Fungible_ReducibleBalance(t *testing.T) {
	tests := []struct {
		name         string
		preservation balancestypes.Preservation
		force        bool
		acc          primitives.AccountInfo
		expectedRes  primitives.Balance
		expectedErr  error
	}{
		{
			name:        "storedMap.Get error",
			expectedErr: expectedErr,
		},
		{
			name: "!Force",
			acc: primitives.AccountInfo{Data: primitives.AccountData{
				Free:     sc.NewU128(100),
				Reserved: sc.NewU128(10),
				Frozen:   sc.NewU128(20),
			}},
			expectedRes: sc.NewU128(90),
		},
		{
			name: "!Force && PreservationPreserve",
			acc: primitives.AccountInfo{Data: primitives.AccountData{
				Free:     sc.NewU128(100),
				Reserved: sc.NewU128(1),
				Frozen:   sc.NewU128(3),
			}},
			preservation: balancestypes.PreservationPreserve,
			expectedRes:  sc.NewU128(98),
		},
		{
			name: "PreservationProtect && HasProviders",
			acc: primitives.AccountInfo{
				Providers: 1,
				Data:      primitives.AccountData{Free: sc.NewU128(100)},
			},

			preservation: balancestypes.PreservationProtect,
			force:        true,
			expectedRes:  sc.NewU128(99),
		},
		{
			name: "PreservationExpendable && !CanDecProviders",
			acc: primitives.AccountInfo{
				Providers: 2,
				Data:      primitives.AccountData{Free: sc.NewU128(100)},
			},
			preservation: balancestypes.PreservationExpendable,
			force:        true,
			expectedRes:  sc.NewU128(100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupFungible()
			mockStoredMap.On("Get", accId).Return(tt.acc, tt.expectedErr)

			result, err := target.reducibleBalance(accId, tt.preservation, tt.force)
			assert.Equal(t, tt.expectedRes, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_Fungible_Transfer(t *testing.T) {
	existentialDeposit := sc.NewU128(1000)
	initialBalance := existentialDeposit.Add(sc.NewU128(10))

	acc1 := make([]byte, 32)
	acc1[0] = 1
	acc2 := make([]byte, 32)
	acc2[0] = 2
	from, _ := primitives.NewAccountId(sc.BytesToSequenceU8(acc1)...)
	to, _ := primitives.NewAccountId(sc.BytesToSequenceU8(acc2)...)

	tests := []struct {
		name                  string
		initialTotalIssuance  primitives.Balance
		initialActiveIssuance primitives.Balance
		initialBalance        primitives.Balance
		transferAmount        primitives.Balance
		preservation          balancestypes.Preservation
		expectedErr           error
	}{
		// Test [`Mutate::transfer`] for a successful transfer.
		//
		// This test verifies that transferring an amount between two accounts with updates the account
		// balances and maintains correct total issuance and active issuance values.
		{
			name:           "transfer success",
			initialBalance: existentialDeposit.Add(sc.NewU128(10)),
			transferAmount: sc.NewU128(3),
		},

		// Test calling [`Mutate::transfer`] with [`Preservation::Expendable`] correctly transfers the
		// entire balance.
		//
		// This test verifies that transferring the entire balance from one account to another with
		// when preservation is expendable updates the account balances and maintains the total
		// issuance and active issuance values.
		{
			name:           "transfer expendable all",
			initialBalance: existentialDeposit.Add(sc.NewU128(10)),
			transferAmount: initialBalance,
		},

		/// Test [`Mutate::transfer`] with [`Preservation::Protect`] and [`Preservation::Preserve`]
		/// transferring the entire balance.
		///
		/// This test verifies that attempting to transfer the entire balance with returns an error when
		/// preservation should not allow it, and the account balances, total issuance, and active
		/// issuance values remain unchanged.
		{
			name:           "transfer protect",
			transferAmount: initialBalance,
			expectedErr:    primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorUnderflow()),
		},
		{
			name:           "transfer preserve",
			transferAmount: initialBalance,
			expectedErr:    primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorUnderflow()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupFungible()
			target.constants.ExistentialDeposit = sc.NewU128(10)
			mockStoredMap.On("Get", from).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: initialBalance, Flags: primitives.DefaultExtraFlags()}}, tt.expectedErr)
			mockStoredMap.On("Get", to).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: initialBalance, Flags: primitives.DefaultExtraFlags()}}, tt.expectedErr)
			mockTotalIssuance.On("Get").Return(tt.initialBalance.Mul(sc.NewU128(2)), nil)
			mockStoredMap.On("TryMutateExistsNoClosure", mock.Anything, mock.Anything).Return(nil)
			mockStoredMap.On("DepositEvent", mock.Anything).Return(nil)
			mockStoredMap.On("IncProviders", mock.Anything).Return(primitives.IncRefStatus(0), nil)

			err := target.transfer(from, to, tt.transferAmount, tt.preservation)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func setupFungible() Module {
	module := setupModule()
	mockStoredMap.On("CanDecProviders", mock.Anything).Return(true, nil)
	return module
}
