package main

import (
	"bytes"
	"github.com/LimeChain/gosemble/primitives/types"
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/secp256k1"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

func Test_Balances_TransferAllowDeath_Success(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	bob, err := ctypes.NewMultiAddressFromHexAccountID(
		"0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22")
	assert.NoError(t, err)

	transferAmount := big.NewInt(0).SetUint64(constants.Dollar)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_allow_death", bob, ctypes.NewUCompact(transferAmount))
	assert.NoError(t, err)

	// Create the extrinsic
	ext := ctypes.NewExtrinsic(call)
	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(testhelpers.ParentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(testhelpers.ParentHash),
		Nonce:              ctypes.NewUCompactFromUInt(0),
		SpecVersion:        ctypes.U32(runtimeVersion.SpecVersion),
		Tip:                ctypes.NewUCompactFromUInt(0),
		TransactionVersion: ctypes.U32(runtimeVersion.TransactionVersion),
	}

	// Set Account Info
	balance, e := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, e)

	keyStorageAccountAlice, aliceAccountInfo := testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)

	// Sign the transaction using Alice's default account
	err = ext.Sign(signature.TestKeyringPairAlice, o)
	assert.NoError(t, err)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = ext.Encode(*encoder)
	assert.NoError(t, err)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), gossamertypes.NewDigest())
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	queryInfo := testhelpers.GetQueryInfo(t, rt, extEnc.Bytes())

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)

	bobHash, _ := common.Blake2b128(bob.AsID[:])
	keyStorageAccountBob := append(testhelpers.KeySystemHash, testhelpers.KeyAccountHash...)
	keyStorageAccountBob = append(keyStorageAccountBob, bobHash...)
	keyStorageAccountBob = append(keyStorageAccountBob, bob.AsID[:]...)
	bytesStorageBob := (*storage).Get(keyStorageAccountBob)

	expectedBobAccountInfo := gossamertypes.AccountInfo{
		Nonce:       0,
		Consumers:   0,
		Producers:   1,
		Sufficients: 0,
		Data: gossamertypes.AccountData{
			Free:       scale.MustNewUint128(transferAmount),
			Reserved:   scale.MustNewUint128(big.NewInt(0)),
			MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
			FreeFrozen: scale.MustNewUint128(types.FlagsNewLogic),
		},
	}

	bobAccountInfo := gossamertypes.AccountInfo{}

	err = scale.Unmarshal(bytesStorageBob, &bobAccountInfo)
	assert.NoError(t, err)

	assert.Equal(t, expectedBobAccountInfo, bobAccountInfo)

	expectedAliceFreeBalance := big.NewInt(0).Sub(
		balance,
		big.NewInt(0).
			Add(transferAmount, queryInfo.PartialFee.ToBigInt()))
	expectedAliceAccountInfo := gossamertypes.AccountInfo{
		Nonce:       1,
		Consumers:   0,
		Producers:   1,
		Sufficients: 0,
		Data: gossamertypes.AccountData{
			Free:       scale.MustNewUint128(expectedAliceFreeBalance),
			Reserved:   scale.MustNewUint128(big.NewInt(0)),
			MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
			FreeFrozen: scale.MustNewUint128(types.FlagsNewLogic),
		},
	}

	bytesAliceStorage := (*storage).Get(keyStorageAccountAlice)
	err = scale.Unmarshal(bytesAliceStorage, &aliceAccountInfo)
	assert.NoError(t, err)

	assert.Equal(t, expectedAliceAccountInfo, aliceAccountInfo)
}

func Test_Balances_TransferAllowDeath_Invalid_InsufficientBalance(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	bob, err := ctypes.NewMultiAddressFromHexAccountID(
		"0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22")
	assert.NoError(t, err)

	transferAmount, e := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, e)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_allow_death", bob, ctypes.NewUCompact(transferAmount))
	assert.NoError(t, err)

	// Create the extrinsic
	ext := ctypes.NewExtrinsic(call)
	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(testhelpers.ParentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(testhelpers.ParentHash),
		Nonce:              ctypes.NewUCompactFromUInt(0),
		SpecVersion:        ctypes.U32(runtimeVersion.SpecVersion),
		Tip:                ctypes.NewUCompactFromUInt(0),
		TransactionVersion: ctypes.U32(runtimeVersion.TransactionVersion),
	}

	// Set Account Info
	balance, e := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, e)
	testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)

	// Sign the transaction using Alice's default account
	err = ext.Sign(signature.TestKeyringPairAlice, o)
	assert.NoError(t, err)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = ext.Encode(*encoder)
	assert.NoError(t, err)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), gossamertypes.NewDigest())
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.Equal(t, testhelpers.ApplyExtrinsicResultTokenErrorFundsUnavailable.Bytes(), res)
}

func Test_Balances_TransferAllowDeath_Invalid_ExistentialDeposit(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	bob, err := ctypes.NewMultiAddressFromHexAccountID(
		"0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22")
	assert.NoError(t, err)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_allow_death", bob, ctypes.NewUCompactFromUInt(1))
	assert.NoError(t, err)

	// Create the extrinsic
	ext := ctypes.NewExtrinsic(call)
	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(testhelpers.ParentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(testhelpers.ParentHash),
		Nonce:              ctypes.NewUCompactFromUInt(0),
		SpecVersion:        ctypes.U32(runtimeVersion.SpecVersion),
		Tip:                ctypes.NewUCompactFromUInt(0),
		TransactionVersion: ctypes.U32(runtimeVersion.TransactionVersion),
	}

	// Set Account Info
	balance, e := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, e)

	testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)

	// Sign the transaction using Alice's default account
	err = ext.Sign(signature.TestKeyringPairAlice, o)
	assert.NoError(t, err)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = ext.Encode(*encoder)
	assert.NoError(t, err)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), gossamertypes.NewDigest())
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultExistentialDepositErr.Bytes(), res)
}

func Test_Balances_TransferAllowDeath_Ecdsa_Signature(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	secpKeypair, err := secp256k1.GenerateKeypair()
	assert.NoError(t, err)

	// Since ECDSA Public Keys are 33 bytes, matching the 32-byte AccountId in the runtime is achieved by blake256(publicKey)
	accountId := blake2b.Sum256(secpKeypair.Public().Encode())

	bob, err := ctypes.NewMultiAddressFromHexAccountID(
		"0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22")
	assert.NoError(t, err)

	transferAmount := big.NewInt(0).SetUint64(constants.Dollar)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_allow_death", bob, ctypes.NewUCompact(transferAmount))
	assert.NoError(t, err)

	// Create the extrinsic
	ext := ctypes.NewExtrinsic(call)
	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(testhelpers.ParentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(testhelpers.ParentHash),
		Nonce:              ctypes.NewUCompactFromUInt(0),
		SpecVersion:        ctypes.U32(runtimeVersion.SpecVersion),
		Tip:                ctypes.NewUCompactFromUInt(0),
		TransactionVersion: ctypes.U32(runtimeVersion.TransactionVersion),
	}

	// Set Account Info
	balance, ok := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, ok)

	keyStorageAccountAlice, aliceAccountInfo := testhelpers.SetStorageAccountInfo(t, storage, accountId[:], balance, 0)

	err = testhelpers.SignExtrinsicSecp256k1(&ext, o, secpKeypair)
	if err != nil {
		panic(err)
	}

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = ext.Encode(*encoder)
	assert.NoError(t, err)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), gossamertypes.NewDigest())
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	queryInfo := testhelpers.GetQueryInfo(t, rt, extEnc.Bytes())

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)

	bobHash, _ := common.Blake2b128(bob.AsID[:])
	keyStorageAccountBob := append(testhelpers.KeySystemHash, testhelpers.KeyAccountHash...)
	keyStorageAccountBob = append(keyStorageAccountBob, bobHash...)
	keyStorageAccountBob = append(keyStorageAccountBob, bob.AsID[:]...)
	bytesStorageBob := (*storage).Get(keyStorageAccountBob)

	expectedBobAccountInfo := gossamertypes.AccountInfo{
		Nonce:       0,
		Consumers:   0,
		Producers:   1,
		Sufficients: 0,
		Data: gossamertypes.AccountData{
			Free:       scale.MustNewUint128(transferAmount),
			Reserved:   scale.MustNewUint128(big.NewInt(0)),
			MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
			FreeFrozen: scale.MustNewUint128(types.FlagsNewLogic),
		},
	}

	bobAccountInfo := gossamertypes.AccountInfo{}

	err = scale.Unmarshal(bytesStorageBob, &bobAccountInfo)
	assert.NoError(t, err)

	assert.Equal(t, expectedBobAccountInfo, bobAccountInfo)

	expectedAliceFreeBalance := big.NewInt(0).Sub(
		balance,
		big.NewInt(0).
			Add(transferAmount, queryInfo.PartialFee.ToBigInt()))
	expectedAliceAccountInfo := gossamertypes.AccountInfo{
		Nonce:       1,
		Consumers:   0,
		Producers:   1,
		Sufficients: 0,
		Data: gossamertypes.AccountData{
			Free:       scale.MustNewUint128(expectedAliceFreeBalance),
			Reserved:   scale.MustNewUint128(big.NewInt(0)),
			MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
			FreeFrozen: scale.MustNewUint128(types.FlagsNewLogic),
		},
	}

	bytesAliceStorage := (*storage).Get(keyStorageAccountAlice)
	err = scale.Unmarshal(bytesAliceStorage, &aliceAccountInfo)
	assert.NoError(t, err)

	assert.Equal(t, expectedAliceAccountInfo, aliceAccountInfo)
}
