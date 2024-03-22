package main

import (
	"math/big"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_AccountNonceApi_account_nonce_Empty(t *testing.T) {
	pubKey := signature.TestKeyringPairAlice.PublicKey

	rt, _ := testhelpers.NewRuntimeInstance(t)

	result, err := rt.Exec("AccountNonceApi_account_nonce", pubKey)
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(0).Bytes(), result)
}

func Test_AccountNonceApi_account_nonce(t *testing.T) {
	pubKey := signature.TestKeyringPairAlice.PublicKey
	rt, storage := testhelpers.NewRuntimeInstance(t)

	nonce := 1

	testhelpers.SetStorageAccountInfo(t, storage, pubKey, big.NewInt(5), 1)

	result, err := rt.Exec("AccountNonceApi_account_nonce", pubKey)
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(nonce).Bytes(), result)
}
