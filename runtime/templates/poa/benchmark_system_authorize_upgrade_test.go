package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemAuthorizeUpgrade(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/system/call_authorize_upgrade_weight.go", func(i *benchmarking.Instance) {
		codeSpecVersion101, err := os.ReadFile(testhelpers.RuntimeWasmSpecVersion101)
		assert.NoError(b, err)
		codeHash := common.MustBlake2bHash(codeSpecVersion101)

		err = i.ExecuteExtrinsic(
			"System.authorize_upgrade",
			primitives.NewRawOriginRoot(),
			codeHash,
		)

		assert.NoError(b, err)
		upgradeAuthorizationBytes := (*i.Storage()).Get(append(testhelpers.KeySystemHash, testhelpers.KeyAuthorizedUpgradeHash...))
		upgradeAuthorization, err := system.DecodeCodeUpgradeAuthorization(bytes.NewBuffer(upgradeAuthorizationBytes))
		assert.NoError(b, err)

		assert.Equal(b, codeHash.ToBytes(), sc.FixedSequenceU8ToBytes(upgradeAuthorization.CodeHash.FixedSequence))
	})
}
