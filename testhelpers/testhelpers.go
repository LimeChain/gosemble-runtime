package testhelpers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/secp256k1"
	"github.com/ChainSafe/gossamer/lib/runtime"
	wazero_runtime "github.com/ChainSafe/gossamer/lib/runtime/wazero"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/ChainSafe/gossamer/pkg/trie"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/balances"
	"github.com/LimeChain/gosemble/frame/session"
	"github.com/LimeChain/gosemble/frame/sudo"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/frame/transaction_payment"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

const RuntimeWasm = "../../../build/runtime.wasm"
const RuntimeWasmSpecVersion101 = "../../../testdata/runtimes/gosemble_poa_template_spec_version_101.wasm"

const (
	SystemIndex sc.U8 = iota
	TimestampIndex
	SessionIndex
	ConsensusAuthoringIndex
	ConsensusFinalizationIndex
	BalancesIndex
	TxPaymentsIndex
	SudoIndex
	TestableIndex = 255
)

// keys from all the modules
var (
	KeySystemHash, _             = common.Twox128Hash([]byte("System"))
	KeyAccountHash, _            = common.Twox128Hash([]byte("Account"))
	KeyAllExtrinsicsLenHash, _   = common.Twox128Hash([]byte("AllExtrinsicsLen"))
	KeyAuraHash, _               = common.Twox128Hash([]byte("Aura"))
	KeyAuthoritiesHash, _        = common.Twox128Hash([]byte("Authorities"))
	KeyAuthorizedUpgradeHash, _  = common.Twox128Hash([]byte("AuthorizedUpgrade"))
	KeyBlockHash, _              = common.Twox128Hash([]byte("BlockHash"))
	KeyCurrentSlotHash, _        = common.Twox128Hash([]byte("CurrentSlot"))
	KeyDigestHash, _             = common.Twox128Hash([]byte("Digest"))
	KeyEventsHash, _             = common.Twox128Hash([]byte("Events"))
	KeyEventCountHash, _         = common.Twox128Hash([]byte("EventCount"))
	KeyExecutionPhaseHash, _     = common.Twox128Hash([]byte("ExecutionPhase"))
	KeyExtrinsicCountHash, _     = common.Twox128Hash([]byte("ExtrinsicCount"))
	KeyExtrinsicIndex            = []byte(":extrinsic_index")
	KeyHeapPages                 = []byte(":heappages")
	KeyExtrinsicDataHash, _      = common.Twox128Hash([]byte("ExtrinsicData"))
	KeyLastRuntimeHash, _        = common.Twox128Hash([]byte("LastRuntimeUpgrade"))
	KeyNumberHash, _             = common.Twox128Hash([]byte("Number"))
	KeyParentHash, _             = common.Twox128Hash([]byte("ParentHash"))
	KeyTimestampHash, _          = common.Twox128Hash([]byte("Timestamp"))
	KeyTimestampNowHash, _       = common.Twox128Hash([]byte("Now"))
	KeyTimestampDidUpdateHash, _ = common.Twox128Hash([]byte("DidUpdate"))
	KeyBlockWeightHash, _        = common.Twox128Hash([]byte("BlockWeight"))
	KeyBalancesHash, _           = common.Twox128Hash([]byte("Balances"))
	KeyTotalIssuanceHash, _      = common.Twox128Hash([]byte("TotalIssuance"))
	KeyTransactionPaymentHash, _ = common.Twox128Hash([]byte("TransactionPayment"))
	KeyNextFeeMultiplierHash, _  = common.Twox128Hash([]byte("NextFeeMultiplier"))
)

// Babe storage keys
var (
	KeyBabeHash, _                     = common.Twox128Hash([]byte("Babe"))
	KeyEpochConfigHash, _              = common.Twox128Hash([]byte("EpochConfig"))
	KeyEpochIndexHash, _               = common.Twox128Hash([]byte("EpochIndex"))
	KeyGenesisSlotHash, _              = common.Twox128Hash([]byte("GenesisSlot"))
	KeyNextRandomnessHash, _           = common.Twox128Hash([]byte("NextRandomness"))
	KeyNextAuthoritiesHash, _          = common.Twox128Hash([]byte("NextAuthorities"))
	KeyNextEpochConfigHash, _          = common.Twox128Hash([]byte("NextEpochConfig"))
	KeyPendingEpochConfigChangeHash, _ = common.Twox128Hash([]byte("PendingEpochConfigChange"))
	KeyRandomnessHash, _               = common.Twox128Hash([]byte("Randomness"))
)

// Grandpa storage keys
var (
	KeyGrandpaHash, _     = common.Twox128Hash([]byte("Grandpa"))
	KeyGrandpaAuthorities = []byte(":grandpa_authorities")
	KeyStalledHash, _     = common.Twox128Hash([]byte("Stalled"))
)

// Session storage keys
var (
	KeySessionHash, _ = common.Twox128Hash([]byte("Session"))
	KeyNextKeys, _    = common.Twox128Hash([]byte("NextKeys"))
	KeyKeyOwner, _    = common.Twox128Hash([]byte("KeyOwner"))
)

// Sudo storage keys
var (
	KeySudoHash, _ = common.Twox128Hash([]byte("Sudo"))
	KeyKeyHash, _  = common.Twox128Hash([]byte("Key"))
)

var (
	ParentHash     = common.MustHexToHash("0x0f6d3477739f8a65886135f58c83ff7c2d4a8300a010dfc8b4c5d65ba37920bb")
	StateRoot      = common.MustHexToHash("0xd9e8bf89bda43fb46914321c371add19b81ff92ad6923e8f189b52578074b073")
	ExtrinsicsRoot = common.MustHexToHash("0x105165e71964828f2b8d1fd89904602cfb9b8930951d87eb249aa2d7c4b51ee7")
	BlockNumber    = uint64(1)
	SealDigest     = gossamertypes.SealDigest{
		ConsensusEngineID: gossamertypes.BabeEngineID,
		// bytes for SealDigest that was created in setupHeaderFile function
		Data: []byte{158, 127, 40, 221, 220, 242, 124, 30, 107, 50, 141, 86, 148, 195, 104, 213, 178, 236, 93, 190,
			14, 65, 42, 225, 201, 143, 136, 213, 59, 228, 216, 80, 47, 172, 87, 31, 63, 25, 201, 202, 175, 40, 26,
			103, 51, 25, 36, 30, 12, 80, 149, 166, 131, 173, 52, 49, 98, 4, 8, 138, 54, 164, 189, 134},
	}
)

var (
	invalidTransactionCallErr              = primitives.NewTransactionValidityError(primitives.NewInvalidTransactionCall())
	invalidTransactionStaleErr             = primitives.NewTransactionValidityError(primitives.NewInvalidTransactionStale())
	invalidTransactionFutureErr            = primitives.NewTransactionValidityError(primitives.NewInvalidTransactionFuture())
	invalidTransactionBadProofErr          = primitives.NewTransactionValidityError(primitives.NewInvalidTransactionBadProof())
	invalidTransactionExhaustsResourcesErr = primitives.NewTransactionValidityError(primitives.NewInvalidTransactionExhaustsResources())
	unknownTransactionNoUnsignedValidator  = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
	invalidTransactionMandatoryValidation  = primitives.NewTransactionValidityError(primitives.NewInvalidTransactionMandatoryValidation())
)

var (
	TransactionValidityResultCallErr, _                = primitives.NewTransactionValidityResult(invalidTransactionCallErr.(primitives.TransactionValidityError))
	TransactionValidityResultStaleErr, _               = primitives.NewTransactionValidityResult(invalidTransactionStaleErr.(primitives.TransactionValidityError))
	TransactionValidityResultFutureErr, _              = primitives.NewTransactionValidityResult(invalidTransactionFutureErr.(primitives.TransactionValidityError))
	TransactionValidityResultExhaustsResourcesErr, _   = primitives.NewTransactionValidityResult(invalidTransactionExhaustsResourcesErr.(primitives.TransactionValidityError))
	TransactionValidityResultNoUnsignedValidatorErr, _ = primitives.NewTransactionValidityResult(unknownTransactionNoUnsignedValidator.(primitives.TransactionValidityError))
	TransactionValidityResultMandatoryValidationErr, _ = primitives.NewTransactionValidityResult(invalidTransactionMandatoryValidation.(primitives.TransactionValidityError))

	dispatchOutcome, _             = primitives.NewDispatchOutcome(nil)
	dispatchOutcomeBadOriginErr, _ = primitives.NewDispatchOutcome(primitives.NewDispatchErrorBadOrigin())

	dispatchOutcomeTokenErrorFundsUnavailable, _ = primitives.NewDispatchOutcome(
		primitives.NewDispatchErrorToken(primitives.NewTokenErrorFundsUnavailable()))
	dispatchOutcomeTokenErrorBelowMinimum, _ = primitives.NewDispatchOutcome(
		primitives.NewDispatchErrorToken(primitives.NewTokenErrorBelowMinimum()))

	dispatchOutcomeSessionNoKeysErr, _ = primitives.NewDispatchOutcome(
		primitives.NewDispatchErrorModule(
			primitives.CustomModuleError{
				Index: SessionIndex,
				Err:   sc.U32(session.ErrorNoKeys),
			}))

	dispatchOutcomeSudoRequireSudoErr, _ = primitives.NewDispatchOutcome(
		primitives.NewDispatchErrorModule(
			primitives.CustomModuleError{
				Index: SudoIndex,
				Err:   sc.U32(sudo.ErrorRequireSudo),
			}))

	ApplyExtrinsicResultOutcome, _                    = primitives.NewApplyExtrinsicResult(dispatchOutcome)
	ApplyExtrinsicResultExhaustsResourcesErr, _       = primitives.NewApplyExtrinsicResult(invalidTransactionExhaustsResourcesErr.(primitives.TransactionValidityError))
	ApplyExtrinsicResultBadOriginErr, _               = primitives.NewApplyExtrinsicResult(dispatchOutcomeBadOriginErr)
	ApplyExtrinsicResultBadProofErr, _                = primitives.NewApplyExtrinsicResult(invalidTransactionBadProofErr.(primitives.TransactionValidityError))
	ApplyExtrinsicResultTokenErrorFundsUnavailable, _ = primitives.NewApplyExtrinsicResult(dispatchOutcomeTokenErrorFundsUnavailable)
	ApplyExtrinsicResultExistentialDepositErr, _      = primitives.NewApplyExtrinsicResult(dispatchOutcomeTokenErrorBelowMinimum)
	ApplyExtrinsicResultSessionNoKeysErr, _           = primitives.NewApplyExtrinsicResult(dispatchOutcomeSessionNoKeysErr)
	ApplyExtrinsicResultSudoRequireSudoErr, _         = primitives.NewApplyExtrinsicResult(dispatchOutcomeSudoRequireSudoErr)
)

var (
	GenesisConfigJson = []byte(
		"{\"system\":{\"code\":\"\"},\"babe\":{\"authorities\":[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"],\"epochConfig\":{\"c\":[1,4],\"allowed_slots\":\"PrimarySlots\"}},\"grandpa\":{\"authorities\":[]},\"balances\":{\"balances\":[[\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",1000000000000000000],[\"5FHneW46xGXgs5mUiveU4sbTyGBzmstUspZC92UhjJM694ty\",1000000000000000000],[\"5FLSigC9HGRKVhB9FiEo4Y3koPsNmBmLJbpXg2mp1hXcS59Y\",1000000000000000000],[\"5DAAnrj7VHTznn2AWBemMuyBwZWs6FNFjdyVXUeYum3PTXFy\",1000000000000000000],[\"5HGjWAeFDfFCWPsjFQdVV2Msvz2XtMktvgocEZcCj68kUMaw\",1000000000000000000],[\"5CiPPseXPECbkjWCa6MnjNokrgYjMqmKndv2rSnekmSK2DjL\",1000000000000000000],[\"5GNJqTPyNqANBkUVMN1LPPrxXnFouWXoe2wNSmmEoLctxiZY\",1000000000000000000],[\"5HpG9w8EBLe5XCrbczpwq5TSXvedjrBGCwqxK1iQ7qUsSWFc\",1000000000000000000],[\"5Ck5SLSHYac6WFt5UZRSsdJjwmpSZq85fd5TRNAdZQVzEAPT\",1000000000000000000],[\"5HKPmK9GYtE1PSLsS1qiYU9xQ9Si1NcEhdeCq9sw5bqu4ns8\",1000000000000000000],[ \"5FCfAonRZgTFrTd9HREEyeJjDpT397KMzizE6T3DvebLFE7n\",1000000000000000000],[\"5CRmqmsiNFExV6VbdmPJViVxrWmkaXXvBrSX8oqBT8R9vmWk\",1000000000000000000]]},\"session\": {\"keys\": [[\"5GNJqTPyNqANBkUVMN1LPPrxXnFouWXoe2wNSmmEoLctxiZY\",\"5GNJqTPyNqANBkUVMN1LPPrxXnFouWXoe2wNSmmEoLctxiZY\",{\"grandpa\":\"5FA9nQDVg267DEd8m1ZypXLBnvN7SFxYwV7ndqSYGiN9TTpu\",\"babe\":\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"}]]},\"transactionPayment\":{\"multiplier\":\"1\"},\"sudo\":{\"key\":\"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\"}}",
	)
)

// TODO:
// implement "Client" type with the helpers functions
// there is also separate "Instance" type for benchmarking that we might get rid of
// and use the "Client" type instead

func NewRuntimeInstance(t *testing.T) (*wazero_runtime.Instance, *runtime.Storage) {
	tt := trie.NewEmptyTrie()
	runtime := wazero_runtime.NewTestInstance(t, RuntimeWasm, wazero_runtime.TestWithTrie(tt))
	return runtime, &runtime.Context.Storage
}

func NewRuntimeInstanceFromCode(t *testing.T, parentRuntime *wazero_runtime.Instance, code []byte) (*wazero_runtime.Instance, *runtime.Storage) {
	cfg := wazero_runtime.Config{
		Storage: parentRuntime.Context.Storage,
	}
	runtime, err := wazero_runtime.NewInstance(code, cfg)
	assert.NoError(t, err)
	return runtime, &runtime.Context.Storage
}

func RuntimeMetadata(t assert.TestingT, instance *wazero_runtime.Instance) *ctypes.Metadata {
	bMetadata, err := instance.Metadata()
	assert.NoError(t, err)

	var decoded []byte
	err = scale.Unmarshal(bMetadata, &decoded)
	assert.NoError(t, err)

	metadata := &ctypes.Metadata{}
	err = codec.Decode(decoded, metadata)
	assert.NoError(t, err)

	return metadata
}

func InitializeBlock(t *testing.T, instance *wazero_runtime.Instance, parentHash, stateRoot, extrinsicsRoot common.Hash, blockNumber uint64) {
	digest := gossamertypes.NewDigest()
	header := gossamertypes.NewHeader(parentHash, stateRoot, extrinsicsRoot, uint(blockNumber), digest)
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = instance.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)
}

func AssertStorageSystemEventCount(t assert.TestingT, storage *runtime.Storage, expected uint32) {
	buffer := &bytes.Buffer{}
	buffer.Write((*storage).Get(append(KeySystemHash, KeyEventCountHash...)))
	storageEventCount, err := sc.DecodeU32(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expected, uint32(storageEventCount))
}

func AssertEmittedBalancesEvent(t assert.TestingT, event sc.U8, buffer *bytes.Buffer) {
	var emitted bool
	eventRecord, err := types.DecodeEventRecord(BalancesIndex, balances.DecodeEvent, buffer)
	assert.NoError(t, err)
	if eventRecord.Event.VaryingData[1] == event {
		emitted = true
	}
	assert.True(t, emitted)
}

func AssertEmittedSystemEvent(t assert.TestingT, event sc.U8, buffer *bytes.Buffer) {
	var emitted bool
	eventRecord, err := types.DecodeEventRecord(SystemIndex, system.DecodeEvent, buffer)
	assert.NoError(t, err)
	if eventRecord.Event.VaryingData[1] == event {
		emitted = true
	}
	assert.True(t, emitted)
}

func AssertEmittedTransactionPaymentEvent(t assert.TestingT, event sc.U8, buffer *bytes.Buffer) {
	var emitted bool
	eventRecord, err := types.DecodeEventRecord(TxPaymentsIndex, transaction_payment.DecodeEvent, buffer)
	assert.NoError(t, err)
	if eventRecord.Event.VaryingData[1] == event {
		emitted = true
	}
	assert.True(t, emitted)
}

func AssertStorageDigestItem(t *testing.T, storage *runtime.Storage, digestItem sc.U8) {
	buffer := bytes.NewBuffer((*storage).Get(append(KeySystemHash, KeyDigestHash...)))
	decodeDigest, err := types.DecodeDigest(buffer)
	assert.NoError(t, err)
	assert.Len(t, decodeDigest.Sequence, 1)
	if decodeDigest.Sequence[0].VaryingData[0] == digestItem {
		assert.True(t, true)
	}
}

func AssertEmittedSudoEvent(t assert.TestingT, event sc.U8, buffer *bytes.Buffer) {
	var emitted bool
	eventRecord, err := types.DecodeEventRecord(SudoIndex, sudo.DecodeEvent, buffer)
	assert.NoError(t, err)
	if eventRecord.Event.VaryingData[1] == event {
		emitted = true
	}
	assert.True(t, emitted)
}

func SetStorageAccountInfo(t *testing.T, storage *runtime.Storage, account []byte, freeBalance *big.Int, nonce uint32) (storageKey []byte, info gossamertypes.AccountInfo) {
	accountInfo := gossamertypes.AccountInfo{
		Nonce:       nonce,
		Consumers:   0,
		Producers:   1,
		Sufficients: 0,
		Data: gossamertypes.AccountData{
			Free:       scale.MustNewUint128(freeBalance),
			Reserved:   scale.MustNewUint128(big.NewInt(0)),
			MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
			FreeFrozen: scale.MustNewUint128(primitives.FlagsNewLogic),
		},
	}

	aliceHash, _ := common.Blake2b128(account)
	keyStorageAccount := append(KeySystemHash, KeyAccountHash...)
	keyStorageAccount = append(keyStorageAccount, aliceHash...)
	keyStorageAccount = append(keyStorageAccount, account...)

	bytesStorage, err := scale.Marshal(accountInfo)
	assert.NoError(t, err)

	err = (*storage).Put(keyStorageAccount, bytesStorage)
	assert.NoError(t, err)

	// Set TotalIssuance as well
	keyTotalIssuance := append(KeyBalancesHash, KeyTotalIssuanceHash...)

	err = (*storage).Put(keyTotalIssuance, sc.NewU128(freeBalance).Bytes())
	assert.NoError(t, err)

	return keyStorageAccount, accountInfo
}

func GetQueryInfo(t *testing.T, instance *wazero_runtime.Instance, extrinsic []byte) primitives.RuntimeDispatchInfo {
	buffer := &bytes.Buffer{}

	buffer.Write(extrinsic)
	err := sc.U32(buffer.Len()).Encode(buffer)
	assert.NoError(t, err)

	bytesRuntimeDispatchInfo, err := instance.Exec("TransactionPaymentApi_query_info", buffer.Bytes())
	assert.NoError(t, err)

	buffer.Reset()
	buffer.Write(bytesRuntimeDispatchInfo)

	dispatchInfo, err := primitives.DecodeRuntimeDispatchInfo(buffer)
	assert.Nil(t, err)

	return dispatchInfo
}

func TimestampExtrinsicBytes(t assert.TestingT, metadata *ctypes.Metadata, time uint64) []byte {
	call, err := ctypes.NewCall(metadata, "Timestamp.set", ctypes.NewUCompactFromUInt(time))
	assert.NoError(t, err)

	expectedExtrinsic := ctypes.NewExtrinsic(call)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = expectedExtrinsic.Encode(*encoder)
	assert.NoError(t, err)

	return extEnc.Bytes()
}

func SignExtrinsicSecp256k1(e *ctypes.Extrinsic, o ctypes.SignatureOptions, keyPair *secp256k1.Keypair) error {
	if e.Type() != ctypes.ExtrinsicVersion4 {
		return fmt.Errorf("unsupported extrinsic version: %v (isSigned: %v, type: %v)", e.Version, e.IsSigned(), e.Type())
	}

	mb, err := codec.Encode(e.Method)
	if err != nil {
		return err
	}

	era := o.Era
	if !o.Era.IsMortalEra {
		era = ctypes.ExtrinsicEra{IsImmortalEra: true}
	}

	payload := ctypes.ExtrinsicPayloadV4{
		ExtrinsicPayloadV3: ctypes.ExtrinsicPayloadV3{
			Method:      mb,
			Era:         era,
			Nonce:       o.Nonce,
			Tip:         o.Tip,
			SpecVersion: o.SpecVersion,
			GenesisHash: o.GenesisHash,
			BlockHash:   o.BlockHash,
		},
		TransactionVersion: o.TransactionVersion,
	}

	b, err := codec.Encode(payload)
	if err != nil {
		return err
	}

	digest := blake2b.Sum256(b)
	signature, err := keyPair.Private().Sign(digest[:])
	if err != nil {
		return err
	}

	signerAddress := blake2b.Sum256(keyPair.Public().Encode())

	signerMultiAddress, err := ctypes.NewMultiAddressFromAccountID(signerAddress[:])
	if err != nil {
		return err
	}

	extSig := ctypes.ExtrinsicSignatureV4{
		Signer:    signerMultiAddress,
		Signature: ctypes.MultiSignature{IsEcdsa: true, AsEcdsa: ctypes.NewEcdsaSignature(signature)},
		Era:       era,
		Nonce:     o.Nonce,
		Tip:       o.Tip,
	}

	e.Signature = extSig

	// mark the extrinsic as signed
	e.Version |= ctypes.ExtrinsicBitSigned

	return nil
}

func AssertSessionNextKeys(t assert.TestingT, storage *runtime.Storage, account []byte, expectedKey []byte) {
	accountHash, _ := common.Twox64(account)
	keySessionNextKeys := append(KeySessionHash, KeyNextKeys...)
	keySessionNextKeys = append(keySessionNextKeys, accountHash...)
	keySessionNextKeys = append(keySessionNextKeys, account...)

	assert.Equal(t, expectedKey, (*storage).Get(keySessionNextKeys))
}

func AssertSessionKeyOwner(t assert.TestingT, storage *runtime.Storage, key primitives.SessionKey, expectedOwner []byte) {
	keyOwnerBytes := key.Bytes()
	keyOwnerHash, _ := common.Twox64(keyOwnerBytes)
	keySessionKeyOwner := append(KeySessionHash, KeyKeyOwner...)
	keySessionKeyOwner = append(keySessionKeyOwner, keyOwnerHash...)
	keySessionKeyOwner = append(keySessionKeyOwner, keyOwnerBytes...)

	fmt.Println(hex.EncodeToString(keySessionKeyOwner))

	assert.Equal(t, expectedOwner, (*storage).Get(keySessionKeyOwner))
}

func AssertSessionEmptyStorage(t assert.TestingT, storage *runtime.Storage, account []byte, key []byte, keyTypeId [4]byte) {
	accountHash, _ := common.Twox64(account)
	keySessionNextKeys := append(KeySessionHash, KeyNextKeys...)
	keySessionNextKeys = append(keySessionNextKeys, accountHash...)
	keySessionNextKeys = append(keySessionNextKeys, account...)

	assert.Nil(t, (*storage).Get(keySessionNextKeys))

	keyOwnerBytes := primitives.NewSessionKey(key, keyTypeId).Bytes()
	keyOwnerHash, _ := common.Twox64(keyOwnerBytes)
	keySessionKeyOwner := append(KeySessionHash, KeyKeyOwner...)
	keySessionKeyOwner = append(keySessionKeyOwner, keyOwnerHash...)
	keySessionKeyOwner = append(keySessionKeyOwner, keyOwnerBytes...)

	assert.Nil(t, (*storage).Get(keySessionKeyOwner))
}

func SetSessionKeysStorage(t assert.TestingT, storage *runtime.Storage, account []byte, key []byte, keyTypeId [4]byte) {
	accountHash, _ := common.Twox64(account)
	keySessionNextKeys := append(KeySessionHash, KeyNextKeys...)
	keySessionNextKeys = append(keySessionNextKeys, accountHash...)
	keySessionNextKeys = append(keySessionNextKeys, account...)

	assert.Nil(t, (*storage).Put(keySessionNextKeys, key))

	keyOwnerBytes := primitives.NewSessionKey(key, keyTypeId).Bytes()
	keyOwnerHash, _ := common.Twox64(keyOwnerBytes)
	keySessionKeyOwner := append(KeySessionHash, KeyKeyOwner...)
	keySessionKeyOwner = append(keySessionKeyOwner, keyOwnerHash...)
	keySessionKeyOwner = append(keySessionKeyOwner, keyOwnerBytes...)

	assert.Nil(t, (*storage).Put(keySessionKeyOwner, account))
}

func GetBabeSlot(t *testing.T, instance *wazero_runtime.Instance, time uint64) uint64 {
	babeConfigurationBytes, err := instance.Exec("BabeApi_configuration", []byte{})
	assert.NoError(t, err)

	babeConfiguration, err := babetypes.DecodeConfiguration(bytes.NewBuffer(babeConfigurationBytes))
	assert.NoError(t, err)

	return time / uint64(babeConfiguration.SlotDuration)
}

func NewBabeDigest(t *testing.T, slot uint64) gossamertypes.Digest {
	primaryDigest := gossamertypes.NewBabePrimaryPreDigest(0, uint64(slot), [32]byte{}, [64]byte{})
	babeDigest := gossamertypes.NewBabeDigest()
	err := babeDigest.SetValue(*primaryDigest)
	assert.NoError(t, err)

	encPreDigestData, err := scale.Marshal(babeDigest)
	assert.NoError(t, err)

	preDigest := gossamertypes.NewBABEPreRuntimeDigest(encPreDigestData)
	digest := gossamertypes.NewDigest()
	err = digest.Add(*preDigest)
	assert.NoError(t, err)

	return digest
}

func GenesisBuild(t *testing.T, instance *wazero_runtime.Instance, genesisConfig []byte) {
	genesisConfigBytes, err := scale.Marshal(genesisConfig)
	assert.NoError(t, err)

	_, err = instance.Exec("GenesisBuilder_build_config", genesisConfigBytes)
	assert.NoError(t, err)
}
