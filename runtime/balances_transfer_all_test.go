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

func Test_Balances_TransferAll_Success_AllowDeath(t *testing.T) {
	rt, storage := newTestRuntime(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := runtimeMetadata(t, rt)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_all", bobAddress, ctypes.NewBool(false))
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
	keyStorageAccountAlice, _ := setStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0, 0, 0)
	// aliceAccountInfo.Providers = 1
	// aliceAccountInfo.Consumers = 1
	// (*storage).Put(keyStorageAccountAlice, aliceAccountInfo.Bytes())

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

	bytesStorageBob := (*storage).Get(keyStorageAccount(bobAccountIdBytes))

	expectedBobAccountInfo := primitives.AccountInfo{
		Providers: 1,
		Data: primitives.AccountData{
			Free: balance.Sub(queryInfo.PartialFee),
			// Free: sc.NewU128(big.NewInt(0).Sub(balanceBigInt, queryInfo.PartialFee.ToBigInt())),
			// Flags: primitives.DefaultExtraFlags(),
			// Reserved:   scale.MustNewUint128(big.NewInt(0)),
			// MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
			// FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
		},
	}

	bobAccountInfo, err := primitives.DecodeAccountInfo(bytes.NewBuffer(bytesStorageBob))
	assert.NoError(t, err)
	assert.Equal(t, expectedBobAccountInfo, bobAccountInfo)

	// expectedAliceAccountInfo := primitives.AccountInfo{
	// 	Nonce:     1,
	// 	Consumers: 0,
	// 	Providers: 0,
	// 	// Providers:   1, // todo after moving UpdateProviders from TryMutateExists to MutateAccount
	// 	Sufficients: 0,
	// 	Data:        primitives.DefaultAccountData(),
	// 	// Data: primitives.AccountData{
	// 	// 	Free:       scale.MustNewUint128(big.NewInt(0)),
	// 	// 	Reserved:   scale.MustNewUint128(big.NewInt(0)),
	// 	// 	MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
	// 	// 	FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
	// 	// },
	// }

	bytesAliceStorage := (*storage).Get(keyStorageAccountAlice)
	assert.Empty(t, bytesAliceStorage)
	// aliceAccountInfo, err = primitives.DecodeAccountInfo(bytes.NewBuffer(bytesAliceStorage))
	// assert.NoError(t, err)

	// assert.Equal(t, primitives.AccountInfo{}, aliceAccountInfo)
}

func Test_Balances_TransferAll_Success_KeepAlive(t *testing.T) {
	rt, storage := newTestRuntime(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := runtimeMetadata(t, rt)

	call, err := ctypes.NewCall(metadata, "Balances.transfer_all", bobAddress, ctypes.NewBool(true))
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

	_, err = rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	// res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	// TODO: remove once tx payments are implemented
	// assert.Equal(t, applyExtrinsicResultKeepAliveErr.Bytes(), res) // todo fix and compare with head

	// TODO: Uncomment once tx payments are implemented, this will be successfully executed,
	// for now it fails due to nothing reserved in account executor
	//assert.Equal(t,
	//	primitives.NewApplyExtrinsicResult(primitives.NewDispatchOutcome(nil)).Bytes(),
	//	res,
	//)

	//bobHash, _ := common.Blake2b128(bob.AsID[:])
	//keyStorageAccountBob := append(keySystemHash, keyAccountHash...)
	//keyStorageAccountBob = append(keyStorageAccountBob, bobHash...)
	//keyStorageAccountBob = append(keyStorageAccountBob, bob.AsID[:]...)
	//bytesStorageBob := storage.Get(keyStorageAccountBob)
	//
	//expectedBobAccountInfo := gossamertypes.AccountInfo{
	//	Nonce:       0,
	//	Consumers:   0,
	//	Producers:   1,
	//	Sufficients: 0,
	//	Data: gossamertypes.AccountData{
	//		Free:       scale.MustNewUint128(mockBalance),
	//		Reserved:   scale.MustNewUint128(big.NewInt(0)),
	//		MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
	//		FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
	//	},
	//}
	//
	//bobAccountInfo := gossamertypes.AccountInfo{}
	//
	//err = scale.Unmarshal(bytesStorageBob, &bobAccountInfo)
	//assert.NoError(t, err)
	//
	//assert.Equal(t, expectedBobAccountInfo, bobAccountInfo)
	//
	//expectedAliceAccountInfo := gossamertypes.AccountInfo{
	//	Nonce:       1,
	//	Consumers:   0,
	//	Producers:   0,
	//	Sufficients: 0,
	//	Data: gossamertypes.AccountData{
	//		Free:       scale.MustNewUint128(big.NewInt(0)),
	//		Reserved:   scale.MustNewUint128(big.NewInt(0)),
	//		MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
	//		FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
	//	},
	//}
	//
	//bytesAliceStorage := storage.Get(keyStorageAccountAlice)
	//err = scale.Unmarshal(bytesAliceStorage, &aliceAccountInfo)
	//assert.NoError(t, err)
	//
	//assert.Equal(t, expectedAliceAccountInfo, aliceAccountInfo)
}
