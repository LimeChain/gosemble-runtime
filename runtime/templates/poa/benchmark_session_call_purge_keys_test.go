package main

import (
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/frame/aura"
	"github.com/LimeChain/gosemble/primitives/types"
	testhelpers "github.com/LimeChain/gosemble/testhelpers"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSessionPurgeKeys(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/session/call_purge_keys_weight.go", func(i *benchmarking.Instance) {
		accountInfo := gossamertypes.AccountInfo{
			Producers: 1,
			Consumers: 1,
			Data: gossamertypes.AccountData{
				Free:       scale.MustNewUint128(big.NewInt(0)),
				Reserved:   scale.MustNewUint128(big.NewInt(0)),
				MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
				FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
			},
		}
		err := i.SetAccountInfo(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)
		testhelpers.SetSessionKeysStorage(b, i.Storage(), signature.TestKeyringPairAlice.PublicKey, bobAddress.AsAddress32[:], aura.KeyTypeId)

		err = i.ExecuteExtrinsic(
			"Session.purge_keys",
			types.NewRawOriginSigned(aliceAccountId),
		)
		assert.NoError(b, err)

		testhelpers.AssertSessionEmptyStorage(b, i.Storage(), signature.TestKeyringPairAlice.PublicKey, bobAddress.AsAddress32[:], aura.KeyTypeId)

		accountInfo, err = i.GetAccountInfo(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Zero(b, accountInfo.Consumers)
	})
}
