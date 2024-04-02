package main

import (
	"bytes"
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_Balances_TransferKeepAlive_Success(t *testing.T) {
	rt, storage := newTestRuntime(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := runtimeMetadata(t, rt)

	transferAmount := big.NewInt(0).SetUint64(constants.Dollar)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_keep_alive", bobAddress, ctypes.NewUCompact(transferAmount))
	assert.NoError(t, err)

	// Create the extrinsic
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

	// Set Account Info
	balanceBigInt, ok := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, ok)
	balance := sc.NewU128(balanceBigInt)

	keyStorageAccountAlice, aliceAccountInfo := setStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 1, 1, 0)

	// Sign the transaction using Alice's default account
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

	queryInfo := getQueryInfo(t, rt, extEnc.Bytes())

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, applyExtrinsicResultOutcome.Bytes(), res)

	expectedBobAccountInfo := primitives.AccountInfo{
		Providers: 1,
		Data: primitives.AccountData{
			Free: sc.NewU128(transferAmount),
		},
	}

	bytesStorageBob := (*storage).Get(keyStorageAccount(bobAccountIdBytes))
	bobAccountInfo, err := primitives.DecodeAccountInfo(bytes.NewBuffer(bytesStorageBob))
	assert.NoError(t, err)

	assert.Equal(t, expectedBobAccountInfo, bobAccountInfo)

	expectedAliceFreeBalance := big.NewInt(0).Sub(
		balanceBigInt,
		big.NewInt(0).
			Add(transferAmount, queryInfo.PartialFee.ToBigInt()))

	bytesAliceStorage := (*storage).Get(keyStorageAccountAlice)
	aliceAccountInfo, err = primitives.DecodeAccountInfo(bytes.NewBuffer(bytesAliceStorage))

	assert.Equal(t, expectedAliceFreeBalance, aliceAccountInfo.Data.Free.ToBigInt())
}
