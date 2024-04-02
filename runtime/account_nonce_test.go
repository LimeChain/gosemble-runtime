package main

import (
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

func Test_AccountNonceApi_account_nonce_Empty(t *testing.T) {
	pubKey := signature.TestKeyringPairAlice.PublicKey

	rt, _ := newTestRuntime(t)

	result, err := rt.Exec("AccountNonceApi_account_nonce", pubKey)
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(0).Bytes(), result)
}

func Test_AccountNonceApi_account_nonce(t *testing.T) {
	pubKey := signature.TestKeyringPairAlice.PublicKey
	rt, storage := newTestRuntime(t)

	nonce := 1

	setStorageAccountInfo(t, storage, pubKey, sc.NewU128(5), 0, 0, 1)

	result, err := rt.Exec("AccountNonceApi_account_nonce", pubKey)
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(nonce).Bytes(), result)
}
