package main

import (
	"bytes"
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/crypto/secp256k1"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

var (
	applyExtrinsicResTokenNoFunds = primitives.ApplyExtrinsicResult{primitives.DispatchOutcome{primitives.NewDispatchErrorToken(primitives.NewTokenErrorNoFunds())}}
)

func Test_Balances_TransferAllowDeath_Success(t *testing.T) {
	rt, storage := newTestRuntime(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := runtimeMetadata(t, rt)

	transferAmount := BalancesExistentialDeposit

	call, err := ctypes.NewCall(metadata, "Balances.transfer_allow_death", bobAddress, ctypes.NewUCompact(transferAmount.ToBigInt()))
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
	keyStorageAccountAlice, _ := setStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, transferAmount.Add(queryInfo.PartialFee), 0, 0, 0)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, applyExtrinsicResultOutcome.Bytes(), res)

	expectedBobAccountInfo := primitives.AccountInfo{
		Providers: 1,
		Data: primitives.AccountData{
			Free: transferAmount,
		},
	}

	bytesStorageBob := (*storage).Get(keyStorageAccount(bobAccountIdBytes))
	bobAccountInfo, err := primitives.DecodeAccountInfo(bytes.NewBuffer(bytesStorageBob))
	assert.NoError(t, err)
	assert.Equal(t, expectedBobAccountInfo, bobAccountInfo)

	bytesAliceStorage := (*storage).Get(keyStorageAccountAlice)
	assert.Empty(t, bytesAliceStorage)
}

func Test_Balances_TransferAllowDeath_Invalid_InsufficientBalance(t *testing.T) {
	rt, storage := newTestRuntime(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := runtimeMetadata(t, rt)

	transferAmount := BalancesExistentialDeposit
	balance := transferAmount.Sub(sc.NewU128(1))
	// transferAmount := big.NewInt(0).SetUint64(constants.Dollar) // todo

	call, err := ctypes.NewCall(metadata, "Balances.transfer_allow_death", bobAddress, ctypes.NewUCompact(transferAmount.ToBigInt()))
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
	setStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0, 0, 0)
	setTotalIssuance(t, storage, transferAmount.Mul(sc.NewU128(2)))

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

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.Equal(t, applyExtrinsicResTokenNoFunds.Bytes(), res)
}

func Test_Balances_TransferAllowDeath_Invalid_ExistentialDeposit(t *testing.T) {
	rt, storage := newTestRuntime(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := runtimeMetadata(t, rt)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_allow_death", bobAddress, ctypes.NewUCompactFromUInt(1))
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

	setStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0, 0, 0)

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

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, applyExtrinsicResTokenNoFunds.Bytes(), res)
}

func Test_Balances_TransferAllowDeath_Ecdsa_Signature(t *testing.T) {
	rt, storage := newTestRuntime(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := runtimeMetadata(t, rt)

	secpKeypair, err := secp256k1.GenerateKeypair()
	assert.NoError(t, err)

	// Since ECDSA Public Keys are 33 bytes, matching the 32-byte AccountId in the runtime is achieved by blake256(publicKey)
	accountId := blake2b.Sum256(secpKeypair.Public().Encode())

	transferAmount := big.NewInt(0).SetUint64(constants.Dollar)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_allow_death", bobAddress, ctypes.NewUCompact(transferAmount))
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

	keyStorageAccountAlice, aliceAccountInfo := setStorageAccountInfo(t, storage, accountId[:], balance, 0, 0, 0)

	err = signExtrinsicSecp256k1(&ext, o, secpKeypair)
	if err != nil {
		panic(err)
	}

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

	bytesStorageBob := (*storage).Get(keyStorageAccount(bobAccountIdBytes))

	expectedBobAccountInfo := primitives.AccountInfo{
		Providers: 1,
		Data: primitives.AccountData{
			Free: sc.NewU128(transferAmount),
		},
	}

	bobAccountInfo, err := primitives.DecodeAccountInfo(bytes.NewBuffer(bytesStorageBob))
	assert.NoError(t, err)
	assert.Equal(t, expectedBobAccountInfo, bobAccountInfo)

	expectedAliceFreeBalance := big.NewInt(0).Sub(
		balanceBigInt,
		big.NewInt(0).
			Add(transferAmount, queryInfo.PartialFee.ToBigInt()))

	bytesAliceStorage := (*storage).Get(keyStorageAccountAlice)
	aliceAccountInfo, err = primitives.DecodeAccountInfo(bytes.NewBuffer(bytesAliceStorage))
	assert.NoError(t, err)

	assert.Equal(t, expectedAliceFreeBalance, aliceAccountInfo.Data.Free.ToBigInt())
}
