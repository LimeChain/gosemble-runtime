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

func BenchmarkSessionSetKeys(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/session/call_set_keys_weight.go", func(i *benchmarking.Instance) {
		accountInfo := gossamertypes.AccountInfo{
			Producers: 1,
			Data: gossamertypes.AccountData{
				Free:       scale.MustNewUint128(big.NewInt(0)),
				Reserved:   scale.MustNewUint128(big.NewInt(0)),
				MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
				FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
			},
		}
		err := i.SetAccountInfo(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"Session.set_keys",
			types.NewRawOriginSigned(aliceAccountId),
			bobAddress.AsAddress32,
			[]byte{0x3, 0x2, 0x1, 0x0},
		)
		assert.NoError(b, err)

		testhelpers.AssertSessionNextKeys(b, i.Storage(), signature.TestKeyringPairAlice.PublicKey, bobAddress.AsAddress32[:])
		testhelpers.AssertSessionKeyOwner(b, i.Storage(), types.NewSessionKey(bobAddress.AsAddress32[:], aura.KeyTypeId), signature.TestKeyringPairAlice.PublicKey)

		accountInfo, err = i.GetAccountInfo(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, uint32(1), accountInfo.Consumers)
	})
}
