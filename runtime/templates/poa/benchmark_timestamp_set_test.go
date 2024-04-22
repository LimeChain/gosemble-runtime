package main

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkTimestampSet(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/timestamp/call_set_weight.go", func(i *benchmarking.Instance) {
		(*i.Storage()).Put(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampNowHash...), sc.U64(0).Bytes())
		(*i.Storage()).DbWhitelistKey(string(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampDidUpdateHash...)))

		now := uint64(dateTime.UnixMilli())

		err := i.ExecuteExtrinsic(
			"Timestamp.set",
			primitives.NewRawOriginNone(),
			ctypes.NewUCompactFromUInt(now),
		)

		assert.NoError(b, err)

		nowStorageValue, err := sc.DecodeU64(bytes.NewBuffer((*i.Storage()).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampNowHash...))))
		assert.NoError(b, err)
		assert.Equal(b, sc.U64(now), nowStorageValue)

		didUpdateStorageValue, err := sc.DecodeBool(bytes.NewBuffer((*i.Storage()).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampDidUpdateHash...))))
		assert.NoError(b, err)
		assert.Equal(b, sc.Bool(true), didUpdateStorageValue)
	})
}
