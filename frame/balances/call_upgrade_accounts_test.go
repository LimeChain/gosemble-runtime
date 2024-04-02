package balances

import (
	"bytes"
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Call_UpgradeAccounts_New(t *testing.T) {
	target := setupCallUpgradeAccounts()
	expected := callUpgradeAccounts{
		Callable: primitives.Callable{
			ModuleId:   target.module.Index,
			FunctionId: functionUpgradeAccountsIndex,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.Sequence[sc.U8]]{}),
		},
		module: target.module,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_UpgradeAccounts_DecodeArgs(t *testing.T) {
	who := sc.Sequence[sc.Sequence[sc.U8]]{sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)}
	target := setupCallUpgradeAccounts()
	buf := &bytes.Buffer{}
	err := who.Encode(buf)
	assert.NoError(t, err)

	call, err := target.DecodeArgs(buf)
	assert.NoError(t, err)
	assert.Equal(t, sc.NewVaryingData(who), call.Args())
}

func Test_Call_UpgradeAccounts_DecodeArgs_Err(t *testing.T) {
	target := setupCallUpgradeAccounts()

	_, err := target.DecodeArgs(&bytes.Buffer{})
	assert.Error(t, err)
}

func Test_Call_UpgradeAccounts_BaseWeight(t *testing.T) {
	target := setupCallUpgradeAccounts()
	assert.Equal(t, callUpgradeAccountsWeight(target.module.dbWeight(), 0), target.BaseWeight())
}

func Test_Call_UpgradeAccounts_WeighData(t *testing.T) {
	target := setupCallUpgradeAccounts()
	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_UpgradeAccounts_ClassifyDispatch(t *testing.T) {
	target := setupCallUpgradeAccounts()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_UpgradeAccounts_PaysFee(t *testing.T) {
	target := setupCallUpgradeAccounts()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_UpgradeAccounts_Dispatch(t *testing.T) {
	tests := []struct {
		name         string
		origin       primitives.RuntimeOrigin
		args         sc.VaryingData
		expectedErr  error
		expectedPays primitives.Pays
	}{
		{
			name:        "error bad origin",
			origin:      primitives.NewRawOriginNone(),
			expectedErr: primitives.NewDispatchErrorBadOrigin(),
		},
		{
			name:        "error invalid arg who",
			origin:      primitives.NewRawOriginSigned(accId),
			args:        sc.NewVaryingData(sc.Empty{}),
			expectedErr: errInvalidAccountIdSequence,
		},
		{
			name:         "error upgradeAccounts",
			origin:       primitives.NewRawOriginSigned(accId),
			args:         sc.NewVaryingData(sc.Sequence[sc.Sequence[sc.U8]]{sc.BytesToSequenceU8([]byte("invalid"))}),
			expectedErr:  errors.New("Address32 should be of size 32"),
			expectedPays: primitives.PaysNo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupCallUpgradeAccounts()

			res, err := target.Dispatch(tt.origin, tt.args)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedPays, res.PaysFee)
		})
	}
}

func Test_Call_UpgradeAccounts_upgradeAccounts(t *testing.T) {
	tests := []struct {
		name                      string
		who                       sc.Sequence[sc.Sequence[sc.U8]]
		expectedFlags             primitives.ExtraFlags
		expectedEnsureUpgradedErr error
		expectedErr               error
		expectedPays              primitives.Pays
		expectedUpgraded          bool
	}{
		{
			name: "empty who",
		},
		{
			name:         "error invalid account id",
			who:          sc.Sequence[sc.Sequence[sc.U8]]{sc.BytesToSequenceU8([]byte("invalid"))},
			expectedErr:  errors.New("Address32 should be of size 32"),
			expectedPays: primitives.PaysNo,
		},
		{
			name:                      "error ensureUpgraded()",
			who:                       sc.Sequence[sc.Sequence[sc.U8]]{sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)},
			expectedEnsureUpgradedErr: expectedErr,
			// expectedErr:               primitives.NewDispatchErrorOther(sc.Str(expectedErr.Error())),
			expectedErr:  expectedErr,
			expectedPays: primitives.PaysNo,
		},
		{
			name:         "upgradeCount >= 90%",
			who:          sc.Sequence[sc.Sequence[sc.U8]]{sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey), sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)},
			expectedPays: primitives.PaysNo,
		},
		{
			name:          "upgradeCount < 90%",
			who:           sc.Sequence[sc.Sequence[sc.U8]]{sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey), sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)},
			expectedFlags: primitives.DefaultExtraFlags(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupCallUpgradeAccounts()

			mockStoredMap.On("Get", accId).Return(primitives.AccountInfo{Data: primitives.AccountData{Flags: tt.expectedFlags}}, tt.expectedEnsureUpgradedErr)
			mockStoredMap.On("TryMutateExistsNoClosure", mock.Anything, mock.Anything).Return(nil) // todo adjust expected mocks for TryMutateExistsNoClosure
			mockStoredMap.On("DepositEvent", mock.Anything)

			pays, err := target.upgradeAccounts(tt.who)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedPays, pays)
		})
	}
}

func setupCallUpgradeAccounts() callUpgradeAccounts {
	return newCallUpgradeAccounts(functionUpgradeAccountsIndex, setupModule())
}
