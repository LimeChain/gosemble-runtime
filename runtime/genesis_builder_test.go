package main

import (
	"bytes"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/grandpa"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
)

func Test_CreateDefaultConfig(t *testing.T) {
	rt, _ := newTestRuntime(t)
	expectedGc := []byte("{\"system\":{},\"session\":{\"keys\":[]},\"aura\":{\"authorities\":[]},\"grandpa\":{\"authorities\":[]},\"balances\":{\"balances\":[]},\"transactionPayment\":{\"multiplier\":\"1\"}}")

	res, err := rt.Exec("GenesisBuilder_create_default_config", []byte{})
	assert.NoError(t, err)

	resDecoded, err := sc.DecodeSequence[sc.U8](bytes.NewBuffer(res))
	assert.Equal(t, expectedGc, sc.SequenceU8ToBytes(resDecoded))
}

func Test_BuildConfig(t *testing.T) {
	rt, storage := newTestRuntime(t)

	gc := []byte("{\"system\":{},\"session\":{\"keys\":[]},\"aura\":{\"authorities\":[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"]},\"grandpa\":{\"authorities\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1]]},\"balances\":{\"balances\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1000000000000000000]]},\"transactionPayment\":{\"multiplier\":\"2\"}}")

	res, err := rt.Exec("GenesisBuilder_build_config", sc.BytesToSequenceU8(gc).Bytes())
	assert.NoError(t, err)
	assert.Equal(t, []byte{0}, res)

	// assert BlockHash
	encBlockNumber, _ := scale.Marshal(uint64(0))
	blockNumHash, _ := common.Twox64(encBlockNumber)
	blockHashKey := append(keySystemHash, keyBlockHash...)
	blockHashKey = append(blockHashKey, blockNumHash...)
	blockHashKey = append(blockHashKey, encBlockNumber...)
	zeroBlockHash := (*storage).Get(blockHashKey)
	expectedBlockHash := types.Blake2bHash69()
	assert.Equal(t, expectedBlockHash.Bytes(), zeroBlockHash)

	// assert ParentHash
	parentHash := (*storage).Get(append(keySystemHash, keyParentHash...))
	assert.Equal(t, expectedBlockHash.Bytes(), parentHash)

	// assert LastRuntimeUpgradeSet
	lrui := (*storage).Get(append(keySystemHash, keyLastRuntimeHash...))
	expectedLrui := types.LastRuntimeUpgradeInfo{SpecVersion: sc.Compact{Number: sc.U32(100)}, SpecName: "node-template"}
	assert.Equal(t, expectedLrui.Bytes(), lrui)

	// assert ExtrinsicIndex
	extrinsicIndex := (*storage).Get(keyExtrinsicIndex)
	expectedExtrinsicIndex := sc.U32(0)
	assert.Equal(t, expectedExtrinsicIndex.Bytes(), extrinsicIndex)

	// assert aura authorities
	auraAuthorities := (*storage).Get(append(keyAuraHash, keyAuthoritiesHash...))
	expectedPubKey := sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)
	expectedAuraAuthorityPubKey, _ := types.NewSr25519PublicKey(expectedPubKey...)
	expectedAuraAuthorities := sc.Sequence[types.Sr25519PublicKey]{expectedAuraAuthorityPubKey}
	assert.Equal(t, expectedAuraAuthorities.Bytes(), auraAuthorities)

	// assert grandpa authorities
	grandpaAuthorities := (*storage).Get(keyGrandpaAuthorities)
	accId, _ := types.NewAccountId(expectedPubKey...)
	authorities := sc.Sequence[types.Authority]{{Id: accId, Weight: sc.U64(1)}}
	expectedGrandpaAuthorities := types.VersionedAuthorityList{AuthorityList: authorities, Version: grandpa.AuthorityVersion}
	assert.Equal(t, expectedGrandpaAuthorities.Bytes(), grandpaAuthorities)

	// assert balance
	accHash, _ := common.Blake2b128(accId.Bytes())
	keyStorageAccount := append(keySystemHash, keyAccountHash...)
	keyStorageAccount = append(keyStorageAccount, accHash...)
	keyStorageAccount = append(keyStorageAccount, accId.Bytes()...)
	accInfo := (*storage).Get(keyStorageAccount)
	expectedBalance := sc.NewU128(uint64(1000000000000000000))
	expectedAccInfo := types.AccountInfo{Data: types.AccountData{Free: expectedBalance}, Providers: 1}
	assert.Equal(t, expectedAccInfo.Bytes(), accInfo)

	// assert total issuance
	totalIssuance := (*storage).Get(append(keyBalancesHash, keyTotalIssuanceHash...))
	assert.Equal(t, expectedBalance.Bytes(), totalIssuance)

	// assert next fee multiplier
	nextFeeMultiplier := (*storage).Get(append(keyTransactionPaymentHash, keyNextFeeMultiplierHash...))
	expectedNextFeeMultiplier := sc.NewU128(2)
	assert.Equal(t, expectedNextFeeMultiplier.Bytes(), nextFeeMultiplier)
}
