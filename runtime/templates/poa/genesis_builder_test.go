package main

import (
	"bytes"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/stretchr/testify/assert"
)

func Test_CreateDefaultConfig(t *testing.T) {
	rt, _ := testhelpers.NewRuntimeInstance(t)
	expectedGc := []byte("{\"system\":{},\"session\":{\"keys\":[]},\"aura\":{\"authorities\":[]},\"grandpa\":{\"authorities\":[]},\"balances\":{\"balances\":[]},\"transactionPayment\":{\"multiplier\":\"1\"},\"sudo\":{\"key\":\"\"}}")

	res, err := rt.Exec("GenesisBuilder_create_default_config", []byte{})
	assert.NoError(t, err)

	resDecoded, err := sc.DecodeSequence[sc.U8](bytes.NewBuffer(res))
	assert.Equal(t, expectedGc, sc.SequenceU8ToBytes(resDecoded))
}

func Test_BuildConfig(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	gc := []byte("{\"system\":{},\"session\":{\"keys\":[]},\"aura\":{\"authorities\":[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"]},\"grandpa\":{\"authorities\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1]]},\"balances\":{\"balances\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1000000000000000000]]},\"transactionPayment\":{\"multiplier\":\"2\"},\"sudo\":{\"key\":\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"}}")

	res, err := rt.Exec("GenesisBuilder_build_config", sc.BytesToSequenceU8(gc).Bytes())
	assert.NoError(t, err)
	assert.Equal(t, []byte{0}, res)

	// assert BlockHash
	encBlockNumber, _ := scale.Marshal(uint64(0))
	blockNumHash, _ := common.Twox64(encBlockNumber)
	blockHashKey := append(testhelpers.KeySystemHash, testhelpers.KeyBlockHash...)
	blockHashKey = append(blockHashKey, blockNumHash...)
	blockHashKey = append(blockHashKey, encBlockNumber...)
	zeroBlockHash := (*storage).Get(blockHashKey)
	expectedBlockHash := types.Blake2bHash69()
	assert.Equal(t, expectedBlockHash.Bytes(), zeroBlockHash)

	// assert ParentHash
	parentHash := (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyParentHash...))
	assert.Equal(t, expectedBlockHash.Bytes(), parentHash)

	// assert LastRuntimeUpgradeSet
	lrui := (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyLastRuntimeHash...))
	expectedLrui := types.LastRuntimeUpgradeInfo{SpecVersion: sc.Compact{Number: sc.U32(100)}, SpecName: "node-template"}
	assert.Equal(t, expectedLrui.Bytes(), lrui)

	// assert ExtrinsicIndex
	extrinsicIndex := (*storage).Get(testhelpers.KeyExtrinsicIndex)
	expectedExtrinsicIndex := sc.U32(0)
	assert.Equal(t, expectedExtrinsicIndex.Bytes(), extrinsicIndex)

	// assert aura authorities
	auraAuthorities := (*storage).Get(append(testhelpers.KeyAuraHash, testhelpers.KeyAuthoritiesHash...))
	expectedPubKey := sc.BytesToSequenceU8(signature.TestKeyringPairAlice.PublicKey)
	expectedAuraAuthorityPubKey, _ := types.NewSr25519PublicKey(expectedPubKey...)
	expectedAuraAuthorities := sc.Sequence[types.Sr25519PublicKey]{expectedAuraAuthorityPubKey}
	assert.Equal(t, expectedAuraAuthorities.Bytes(), auraAuthorities)

	// assert grandpa authorities
	grandpaAuthorities := (*storage).Get(testhelpers.KeyGrandpaAuthorities)
	accId, _ := types.NewAccountId(expectedPubKey...)
	authorities := sc.Sequence[types.Authority]{{Id: accId, Weight: sc.U64(1)}}
	assert.Equal(t, authorities.Bytes(), grandpaAuthorities)

	// assert balance
	accHash, _ := common.Blake2b128(accId.Bytes())
	keyStorageAccount := append(testhelpers.KeySystemHash, testhelpers.KeyAccountHash...)
	keyStorageAccount = append(keyStorageAccount, accHash...)
	keyStorageAccount = append(keyStorageAccount, accId.Bytes()...)
	accInfo := (*storage).Get(keyStorageAccount)
	expectedBalance := sc.NewU128(uint64(1000000000000000000))
	expectedAccInfo := types.AccountInfo{Data: types.AccountData{Free: expectedBalance, Flags: types.DefaultExtraFlags}, Providers: 1}
	assert.Equal(t, expectedAccInfo.Bytes(), accInfo)

	// assert total issuance
	totalIssuance := (*storage).Get(append(testhelpers.KeyBalancesHash, testhelpers.KeyTotalIssuanceHash...))
	assert.Equal(t, expectedBalance.Bytes(), totalIssuance)

	// assert next fee multiplier
	nextFeeMultiplier := (*storage).Get(append(testhelpers.KeyTransactionPaymentHash, testhelpers.KeyNextFeeMultiplierHash...))
	expectedNextFeeMultiplier := sc.NewU128(2)
	assert.Equal(t, expectedNextFeeMultiplier.Bytes(), nextFeeMultiplier)

	// assert sudo key
	assert.Equal(t, aliceAddress.AsID.ToBytes(), (*storage).Get(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...)))
}
