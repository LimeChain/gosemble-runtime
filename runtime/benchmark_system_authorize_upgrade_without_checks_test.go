package main

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemAuthorizeUpgradeWithoutChecks(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/system/call_authorize_upgrade_without_checks_weight.go", func(i *benchmarking.Instance) {
		err := i.ExecuteExtrinsic(
			"System.authorize_upgrade_without_checks",
			primitives.NewRawOriginRoot(),
			codeHash,
		)

		assert.NoError(b, err)
		upgradeAuthorizationBytes := (*i.Storage()).Get(append(keySystemHash, keyAuthorizedUpgradeHash...))
		upgradeAuthorization, err := system.DecodeCodeUpgradeAuthorization(bytes.NewBuffer(upgradeAuthorizationBytes))
		assert.NoError(b, err)

		assert.Equal(b, codeHash.ToBytes(), sc.FixedSequenceU8ToBytes(upgradeAuthorization.CodeHash.FixedSequence))
	})
}
