package main

import (
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_Grandpa_Authorities_Empty(t *testing.T) {
	rt, _ := testhelpers.NewRuntimeInstance(t)

	result, err := rt.Exec("GrandpaApi_grandpa_authorities", []byte{})
	assert.NoError(t, err)

	assert.Equal(t, []byte{0}, result)
}

func Test_Grandpa_Authorities(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	pubKey1 := common.MustHexToBytes("0x88dc3417d5058ec4b4503e0c12ea1a0a89be200fe98922423d4334014fa6b0ee")
	pubKey2 := common.MustHexToBytes("0x88dc3417d5058ec4b4503e0c12ea1a0a89be200fe98922423d4334014fa6b0ef")
	weight := sc.U64(1)

	storageAuthorities := sc.Sequence[types.Authority]{
		{
			Id:     types.AccountId{FixedSequence: sc.BytesToFixedSequenceU8(pubKey1)},
			Weight: weight,
		},
		{
			Id:     types.AccountId{FixedSequence: sc.BytesToFixedSequenceU8(pubKey2)},
			Weight: weight,
		},
	}

	err := (*storage).Put([]byte(":grandpa_authorities"), storageAuthorities.Bytes())
	assert.NoError(t, err)

	result, err := rt.Exec("GrandpaApi_grandpa_authorities", []byte{})
	assert.NoError(t, err)

	assert.Equal(t, storageAuthorities.Bytes(), result)
}
