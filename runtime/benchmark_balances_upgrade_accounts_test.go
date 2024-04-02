package main

import (
	"testing"

	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkBalancesUpgradeAccounts(b *testing.B) {
	size, err := benchmarking.NewLinear("size", 1, 1000)
	assert.NoError(b, err)

	benchmarking.RunDispatchCall(b, "../frame/balances/call_upgrade_accounts_weight.go", func(instance *benchmarking.Instance) {
		who := make(sc.Sequence[sc.Sequence[sc.U8]], size.Value())
		for i, _ := range who {
			kp, err := sr25519.GenerateKeypair()
			assert.NoError(b, err)

			accountInfo := primitives.AccountInfo{
				Providers: 1,
				Data: primitives.AccountData{
					Free:     BalancesExistentialDeposit.Mul(sc.NewU128(10)),
					Reserved: BalancesExistentialDeposit.Mul(sc.NewU128(10)),
				},
			}
			err = instance.SetAccountInfoNew(kp.Public().Encode(), accountInfo)
			assert.NoError(b, err)

			accountInfo, err = instance.GetAccountInfoNew(kp.Public().Encode())
			assert.Equal(b, sc.U32(1), accountInfo.Providers)
			assert.Zero(b, accountInfo.Consumers)
			assert.False(b, accountInfo.Data.Flags.IsNewLogic())

			who[i] = sc.BytesToSequenceU8(kp.Public().Encode())
		}

		err := instance.ExecuteExtrinsic(
			"Balances.upgrade_accounts",
			primitives.NewRawOriginSigned(aliceAccountId),
			who,
		)
		assert.NoError(b, err)

		for _, accId := range who {
			accountInfo, err := instance.GetAccountInfoNew(sc.SequenceU8ToBytes(accId))
			assert.NoError(b, err)
			assert.Equal(b, sc.U32(1), accountInfo.Providers)
			assert.Equal(b, sc.U32(1), accountInfo.Consumers)
			assert.True(b, accountInfo.Data.Flags.IsNewLogic())
		}
	}, size)
}
