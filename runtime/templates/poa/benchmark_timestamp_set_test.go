package main

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkTimestampSet(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/timestamp/call_set_weight.go", func(i *benchmarking.Instance) {
		// arrange
		(*i.Storage()).Put(append(keyTimestampHash, keyTimestampNowHash...), sc.U64(0).Bytes())
		(*i.Storage()).DbWhitelistKey(string(append(keyTimestampHash, keyTimestampDidUpdateHash...)))

		now := uint64(dateTime.UnixMilli())

		// act
		err := i.ExecuteExtrinsic(
			"Timestamp.set",
			primitives.NewRawOriginNone(),
			ctypes.NewUCompactFromUInt(now),
		)

		// assert
		assert.NoError(b, err)

		nowStorageValue, err := sc.DecodeU64(bytes.NewBuffer((*i.Storage()).Get(append(keyTimestampHash, keyTimestampNowHash...))))
		assert.NoError(b, err)
		assert.Equal(b, sc.U64(now), nowStorageValue)

		didUpdateStorageValue, err := sc.DecodeBool(bytes.NewBuffer((*i.Storage()).Get(append(keyTimestampHash, keyTimestampDidUpdateHash...))))
		assert.NoError(b, err)
		assert.Equal(b, sc.Bool(true), didUpdateStorageValue)
	})
}
