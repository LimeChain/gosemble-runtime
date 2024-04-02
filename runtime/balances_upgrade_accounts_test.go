package main

import (
	"bytes"
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_Balances_UpgradeAccounts(t *testing.T) {
	rt, storage := newTestRuntime(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := runtimeMetadata(t, rt)
	call, err := ctypes.NewCall(metadata, "Balances.upgrade_accounts", sc.Sequence[sc.Sequence[sc.U8]]{sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)})
	assert.NoError(t, err)

	ext := ctypes.NewExtrinsic(call)
	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(parentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(parentHash),
		Nonce:              ctypes.NewUCompactFromUInt(0),
		SpecVersion:        ctypes.U32(runtimeVersion.SpecVersion),
		Tip:                ctypes.NewUCompactFromUInt(0),
		TransactionVersion: ctypes.U32(runtimeVersion.TransactionVersion),
	}

	balanceBigInt, ok := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, ok)
	balance := sc.NewU128(balanceBigInt)

	keyStorageAccountAlice, aliceAccountInfo := setStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0, 0, 0)
	assert.Zero(t, aliceAccountInfo.Data.Flags)
	assert.False(t, aliceAccountInfo.Data.Flags.IsNewLogic())

	err = ext.Sign(signature.TestKeyringPairAlice, o)
	assert.NoError(t, err)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = ext.Encode(*encoder)
	assert.NoError(t, err)

	header := gossamertypes.NewHeader(parentHash, stateRoot, extrinsicsRoot, uint(blockNumber), gossamertypes.NewDigest())
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	_, err = rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	bytesAliceStorage := (*storage).Get(keyStorageAccountAlice)
	aliceAccountInfo, err = primitives.DecodeAccountInfo(bytes.NewBuffer(bytesAliceStorage))
	assert.NoError(t, err)

	assert.True(t, aliceAccountInfo.Data.Flags.IsNewLogic())
}
