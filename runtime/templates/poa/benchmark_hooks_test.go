package main

import (
	"testing"

	"github.com/ChainSafe/gossamer/lib/runtime"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
)

func BenchmarkOnInitialize(b *testing.B) {
	auraCurrentSlot := sc.U64(1)
	auraNewSlot := sc.U64(2)
	auraAuthorityPubKey, _ := types.NewSr25519PublicKey(sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)...)
	auraAuthorities := sc.Sequence[types.Sr25519PublicKey]{auraAuthorityPubKey}
	digest := types.NewDigest(sc.Sequence[types.DigestItem]{
		types.NewDigestItemPreRuntime(
			sc.BytesToFixedSequenceU8([]byte{'a', 'u', 'r', 'a'}),
			sc.BytesToSequenceU8(auraNewSlot.Bytes()),
		),
	})

	setup := func(storage *runtime.Storage) {
		(*storage).Put(append(keySystemHash, keyDigestHash...), digest.Bytes())
		(*storage).Put(append(keyAuraHash, keyCurrentSlotHash...), auraCurrentSlot.Bytes())
		(*storage).Put(append(keyAuraHash, keyAuthoritiesHash...), auraAuthorities.Bytes())
	}

	validate := func(storage *runtime.Storage) {
		assert.Equal(b, sc.U64(2).Bytes(), (*storage).Get(append(keyAuraHash, keyCurrentSlotHash...)))
	}

	benchmarking.RunHook(b, "on_initialize", setup, validate)
}

func BenchmarkOnRuntimeUpgrade(b *testing.B) {
	setup := func(storage *runtime.Storage) {}
	validate := func(storage *runtime.Storage) {}

	benchmarking.RunHook(b, "on_runtime_upgrade", setup, validate)
}

func BenchmarkOnFinalize(b *testing.B) {
	key := append(keyTimestampHash, keyTimestampDidUpdateHash...)

	setup := func(storage *runtime.Storage) {
		(*storage).Put(key, sc.Bool(true).Bytes())
	}

	validate := func(storage *runtime.Storage) {
		assert.Equal(b, []byte(nil), (*storage).Get(key))
	}

	benchmarking.RunHook(b, "on_finalize", setup, validate)
}

func BenchmarkOnIdle(b *testing.B) {
	setup := func(storage *runtime.Storage) {}
	validate := func(storage *runtime.Storage) {}

	benchmarking.RunHook(b, "on_idle", setup, validate)
}
