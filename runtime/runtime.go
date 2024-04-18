/*
Targets WebAssembly MVP
*/
package main

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/api/account_nonce"
	apiAura "github.com/LimeChain/gosemble/api/aura"
	"github.com/LimeChain/gosemble/api/benchmarking"
	blockbuilder "github.com/LimeChain/gosemble/api/block_builder"
	"github.com/LimeChain/gosemble/api/collect_collation_info"
	"github.com/LimeChain/gosemble/api/core"
	genesisbuilder "github.com/LimeChain/gosemble/api/genesis_builder"
	apiGrandpa "github.com/LimeChain/gosemble/api/grandpa"
	"github.com/LimeChain/gosemble/api/metadata"
	"github.com/LimeChain/gosemble/api/offchain_worker"
	"github.com/LimeChain/gosemble/api/parachain"
	"github.com/LimeChain/gosemble/api/session_keys"
	taggedtransactionqueue "github.com/LimeChain/gosemble/api/tagged_transaction_queue"
	apiTxPayments "github.com/LimeChain/gosemble/api/transaction_payment"
	apiTxPaymentsCall "github.com/LimeChain/gosemble/api/transaction_payment_call"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/execution/extrinsic"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/aura"
	aura_ext "github.com/LimeChain/gosemble/frame/aura-ext"
	"github.com/LimeChain/gosemble/frame/balances"
	"github.com/LimeChain/gosemble/frame/executive"
	"github.com/LimeChain/gosemble/frame/grandpa"
	"github.com/LimeChain/gosemble/frame/parachain_info"
	"github.com/LimeChain/gosemble/frame/parachain_system"
	"github.com/LimeChain/gosemble/frame/session"
	"github.com/LimeChain/gosemble/frame/system"
	sysExtensions "github.com/LimeChain/gosemble/frame/system/extensions"
	tm "github.com/LimeChain/gosemble/frame/testable"
	"github.com/LimeChain/gosemble/frame/timestamp"
	"github.com/LimeChain/gosemble/frame/transaction_payment"
	txExtensions "github.com/LimeChain/gosemble/frame/transaction_payment/extensions"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	AuraMaxAuthorities = 100
)

const (
	BalancesMaxLocks    = 50
	BalancesMaxReserves = 50
)

const (
	TimestampMinimumPeriod = 1 * 1_000 // 1 second
)

const (
	MilliSecsPerBlock = 2_000
	SlotDuration      = MilliSecsPerBlock
	Minutes           = 60_000 / MilliSecsPerBlock
	Hours             = Minutes * 60
	Days              = Hours * 24
)

var (
	BalancesExistentialDeposit = sc.NewU128(1 * constants.Dollar)
)

var (
	DbWeight = constants.RocksDbWeight
)

var (
	OperationalFeeMultiplier                        = sc.U8(5)
	WeightToFee              primitives.WeightToFee = primitives.IdentityFee{}
	LengthToFee              primitives.WeightToFee = primitives.IdentityFee{}
)

var (
	Period sc.U64 = 6 * Hours
	Offset sc.U64 = 0
)

// Parachain
const (
	RelayChainSlotDurationMillis = 6_000
	BlockProcessingVelocity      = 1
	UnincludedSegmentCapacity    = 1
)

const (
	SystemIndex sc.U8 = iota
	ParachainSystem
	TimestampIndex
	SessionIndex
	ParachainInfo
	AuraIndex
	AuraExtIndex
	GrandpaIndex
	BalancesIndex
	TxPaymentsIndex
	TestableIndex = 255
)

var (
	// RuntimeVersion contains the version identifiers of the Runtime.
	RuntimeVersion = &primitives.RuntimeVersion{
		SpecName:           sc.Str(constants.SpecName),
		ImplName:           sc.Str(constants.ImplName),
		AuthoringVersion:   sc.U32(constants.AuthoringVersion),
		SpecVersion:        sc.U32(constants.SpecVersion),
		ImplVersion:        sc.U32(constants.ImplVersion),
		TransactionVersion: sc.U32(constants.TransactionVersion),
		StateVersion:       sc.U8(constants.StateVersion),
	}

	// Block default values used in module initialization.
	blockWeights, blockLength = initializeBlockDefaults()

	maxConsumers sc.U32 = 16
)

var (
	logger      = log.NewLogger()
	mdGenerator = primitives.NewMetadataTypeGenerator()
	// Modules contains all the modules used by the runtime.
	modules = initializeModules()
	extra   = newSignedExtra()
	decoder = types.NewRuntimeDecoder(modules, extra, logger)
)

func initializeBlockDefaults() (primitives.BlockWeights, primitives.BlockLength) {
	weights, err := system.WithSensibleDefaults(constants.MaximumBlockWeight, constants.NormalDispatchRatio)
	if err != nil {
		logger.Critical(err.Error())
	}

	length, err := system.MaxWithNormalRatio(constants.FiveMbPerBlockPerExtrinsic, constants.NormalDispatchRatio)
	if err != nil {
		logger.Critical(err.Error())
	}

	return weights, length
}

func initializeModules() []primitives.Module {
	systemModule := system.New(
		SystemIndex,
		system.NewConfig(primitives.BlockHashCount{U32: sc.U32(constants.BlockHashCount)}, blockWeights, blockLength, DbWeight, RuntimeVersion, maxConsumers),
		mdGenerator,
		logger,
	)

	parachainSystemModule := parachain_system.New(ParachainSystem, parachain_system.NewConfig(DbWeight))
	parachainInfoModule := parachain_info.New(ParachainInfo)

	auraModule := aura.New(
		AuraIndex,
		aura.NewConfig(
			primitives.PublicKeySr25519,
			DbWeight,
			TimestampMinimumPeriod,
			AuraMaxAuthorities,
			false,
			systemModule.StorageDigest,
			systemModule,
			nil,
		),
		mdGenerator,
		logger,
	)

	auraExtModule := aura_ext.New(AuraExtIndex, aura_ext.NewConfig(DbWeight, RelayChainSlotDurationMillis, BlockProcessingVelocity, UnincludedSegmentCapacity), auraModule, logger)

	timestampModule := timestamp.New(
		TimestampIndex,
		timestamp.NewConfig(auraModule, DbWeight, TimestampMinimumPeriod),
		mdGenerator,
	)

	grandpaModule := grandpa.New(GrandpaIndex, logger, mdGenerator)

	handler := session.NewHandler([]session.OneSessionHandler{auraModule})

	periodicSession := session.NewPeriodicSessions(Period, Offset)
	sessionModule := session.New(
		SessionIndex,
		session.NewConfig(DbWeight, blockWeights, systemModule, periodicSession, handler, session.DefaultManager{}),
		mdGenerator,
		logger)

	balancesModule := balances.New(
		BalancesIndex,
		balances.NewConfig(DbWeight, BalancesMaxLocks, BalancesMaxReserves, BalancesExistentialDeposit, systemModule),
		logger,
		mdGenerator,
	)

	tpmModule := transaction_payment.New(
		TxPaymentsIndex,
		transaction_payment.NewConfig(OperationalFeeMultiplier, WeightToFee, LengthToFee, blockWeights),
		mdGenerator,
	)

	testableModule := tm.New(TestableIndex, mdGenerator)

	return []primitives.Module{
		systemModule,
		parachainSystemModule,
		timestampModule,
		sessionModule,
		parachainInfoModule,
		auraModule,
		auraExtModule,
		grandpaModule,
		balancesModule,
		tpmModule,
		testableModule,
	}
}

func newSignedExtra() primitives.SignedExtra {
	systemModule := primitives.MustGetModule(SystemIndex, modules).(system.Module)
	balancesModule := primitives.MustGetModule(BalancesIndex, modules).(balances.Module)
	txPaymentModule := primitives.MustGetModule(TxPaymentsIndex, modules).(transaction_payment.Module)

	extras := []primitives.SignedExtension{
		sysExtensions.NewCheckNonZeroAddress(),
		sysExtensions.NewCheckSpecVersion(systemModule),
		sysExtensions.NewCheckTxVersion(systemModule),
		sysExtensions.NewCheckGenesis(systemModule),
		sysExtensions.NewCheckMortality(systemModule),
		sysExtensions.NewCheckNonce(systemModule),
		sysExtensions.NewCheckWeight(systemModule),
		txExtensions.NewChargeTransactionPayment(systemModule, txPaymentModule, balancesModule),
	}

	return primitives.NewSignedExtra(extras, mdGenerator)
}

func runtimeApi() types.RuntimeApi {
	runtimeExtrinsic := extrinsic.New(modules, extra, mdGenerator, logger)
	systemModule := primitives.MustGetModule(SystemIndex, modules).(system.Module)
	auraModule := primitives.MustGetModule(AuraIndex, modules).(aura.Module)
	grandpaModule := primitives.MustGetModule(GrandpaIndex, modules).(grandpa.Module)
	txPaymentsModule := primitives.MustGetModule(TxPaymentsIndex, modules).(transaction_payment.Module)
	parachainSystemModule := primitives.MustGetModule(ParachainSystem, modules).(parachain_system.Module)

	executiveModule := executive.New(
		systemModule,
		runtimeExtrinsic,
		hooks.DefaultOnRuntimeUpgrade{},
		logger,
	)
	// TODO: Aura ext uses executive

	sessions := []primitives.Session{
		auraModule,
		grandpaModule,
	}

	coreApi := core.New(executiveModule, decoder, RuntimeVersion, mdGenerator, logger)
	blockBuilderApi := blockbuilder.New(runtimeExtrinsic, executiveModule, decoder, mdGenerator, logger)
	taggedTxQueueApi := taggedtransactionqueue.New(executiveModule, decoder, mdGenerator, logger)
	auraApi := apiAura.New(auraModule, logger)
	grandpaApi := apiGrandpa.New(grandpaModule, logger)
	accountNonceApi := account_nonce.New(systemModule, logger)
	txPaymentsApi := apiTxPayments.New(decoder, txPaymentsModule, logger)
	txPaymentsCallApi := apiTxPaymentsCall.New(decoder, txPaymentsModule, logger)
	sessionKeysApi := session_keys.New(sessions, logger)
	offchainWorkerApi := offchain_worker.New(executiveModule, logger)
	genesisBuilderApi := genesisbuilder.New(modules, logger)
	collectCollationInfoApi := collect_collation_info.New(parachainSystemModule, logger)
	// TODO: Check if this API module has metadata as well
	parachainApi := parachain.New(parachainSystemModule, decoder, logger)

	metadataApi := metadata.New(
		runtimeExtrinsic,
		[]primitives.RuntimeApiModule{
			coreApi,
			blockBuilderApi,
			taggedTxQueueApi,
			auraApi,
			grandpaApi,
			accountNonceApi,
			txPaymentsApi,
			txPaymentsCallApi,
			sessionKeysApi,
			offchainWorkerApi,
			collectCollationInfoApi,
		},
		logger,
		mdGenerator,
	)

	apis := []primitives.ApiModule{
		coreApi,
		blockBuilderApi,
		taggedTxQueueApi,
		metadataApi,
		auraApi,
		grandpaApi,
		accountNonceApi,
		txPaymentsApi,
		txPaymentsCallApi,
		sessionKeysApi,
		offchainWorkerApi,
		genesisBuilderApi,
		collectCollationInfoApi,
		parachainApi,
	}

	runtimeApi := types.NewRuntimeApi(apis, logger)

	RuntimeVersion.SetApis(runtimeApi.Items())

	return runtimeApi
}

//go:export Core_version
func CoreVersion(_ int32, _ int32) int64 {
	return runtimeApi().
		Module(core.ApiModuleName).(core.Core).
		Version()
}

//go:export Core_initialize_block
func CoreInitializeBlock(dataPtr int32, dataLen int32) int64 {
	runtimeApi().
		Module(core.ApiModuleName).(core.Core).
		InitializeBlock(dataPtr, dataLen)

	return 0
}

//go:export Core_execute_block
func CoreExecuteBlock(dataPtr int32, dataLen int32) int64 {
	runtimeApi().Module(core.ApiModuleName).(core.Core).
		ExecuteBlock(dataPtr, dataLen)

	return 0
}

//go:export BlockBuilder_apply_extrinsic
func BlockBuilderApplyExtrinsic(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(blockbuilder.ApiModuleName).(blockbuilder.BlockBuilder).
		ApplyExtrinsic(dataPtr, dataLen)
}

//go:export BlockBuilder_finalize_block
func BlockBuilderFinalizeBlock(_, _ int32) int64 {
	return runtimeApi().
		Module(blockbuilder.ApiModuleName).(blockbuilder.BlockBuilder).
		FinalizeBlock()
}

//go:export BlockBuilder_inherent_extrinsics
func BlockBuilderInherentExtrinsics(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(blockbuilder.ApiModuleName).(blockbuilder.BlockBuilder).
		InherentExtrinsics(dataPtr, dataLen)
}

//go:export BlockBuilder_check_inherents
func BlockBuilderCheckInherents(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(blockbuilder.ApiModuleName).(blockbuilder.BlockBuilder).
		CheckInherents(dataPtr, dataLen)
}

//go:export TaggedTransactionQueue_validate_transaction
func TaggedTransactionQueueValidateTransaction(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(taggedtransactionqueue.ApiModuleName).(taggedtransactionqueue.TaggedTransactionQueue).
		ValidateTransaction(dataPtr, dataLen)
}

//go:export AuraApi_slot_duration
func AuraApiSlotDuration(_, _ int32) int64 {
	return runtimeApi().
		Module(apiAura.ApiModuleName).(apiAura.Module).
		SlotDuration()
}

//go:export AuraApi_authorities
func AuraApiAuthorities(_, _ int32) int64 {
	return runtimeApi().
		Module(apiAura.ApiModuleName).(apiAura.Module).
		Authorities()
}

//go:export AccountNonceApi_account_nonce
func AccountNonceApiAccountNonce(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(account_nonce.ApiModuleName).(account_nonce.Module).
		AccountNonce(dataPtr, dataLen)
}

//go:export TransactionPaymentApi_query_info
func TransactionPaymentApiQueryInfo(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(apiTxPayments.ApiModuleName).(apiTxPayments.Module).
		QueryInfo(dataPtr, dataLen)
}

//go:export TransactionPaymentApi_query_fee_details
func TransactionPaymentApiQueryFeeDetails(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(apiTxPayments.ApiModuleName).(apiTxPayments.Module).
		QueryFeeDetails(dataPtr, dataLen)
}

//go:export TransactionPaymentCallApi_query_call_info
func TransactionPaymentCallApiQueryCallInfo(dataPtr int32, dataLan int32) int64 {
	return runtimeApi().
		Module(apiTxPaymentsCall.ApiModuleName).(apiTxPaymentsCall.Module).
		QueryCallInfo(dataPtr, dataLan)
}

//go:export TransactionPaymentCallApi_query_call_fee_details
func TransactionPaymentCallApiQueryCallFeeDetails(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(apiTxPaymentsCall.ApiModuleName).(apiTxPaymentsCall.Module).
		QueryCallFeeDetails(dataPtr, dataLen)
}

//go:export Metadata_metadata
func Metadata(_, _ int32) int64 {
	mdGenerator.ClearMetadata()
	return runtimeApi().
		Module(metadata.ApiModuleName).(metadata.Module).
		Metadata()
}

//go:export Metadata_metadata_at_version
func MetadataAtVersion(dataPtr int32, dataLen int32) int64 {
	mdGenerator.ClearMetadata()
	return runtimeApi().
		Module(metadata.ApiModuleName).(metadata.Module).
		MetadataAtVersion(dataPtr, dataLen)
}

//go:export Metadata_metadata_versions
func MetadataVersions(_, _ int32) int64 {
	return runtimeApi().
		Module(metadata.ApiModuleName).(metadata.Module).
		MetadataVersions()
}

//go:export SessionKeys_generate_session_keys
func SessionKeysGenerateSessionKeys(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(session_keys.ApiModuleName).(session_keys.Module).
		GenerateSessionKeys(dataPtr, dataLen)
}

//go:export SessionKeys_decode_session_keys
func SessionKeysDecodeSessionKeys(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(session_keys.ApiModuleName).(session_keys.Module).
		DecodeSessionKeys(dataPtr, dataLen)
}

//go:export GrandpaApi_grandpa_authorities
func GrandpaApiAuthorities(_, _ int32) int64 {
	return runtimeApi().
		Module(apiGrandpa.ApiModuleName).(apiGrandpa.Module).
		Authorities()
}

//go:export OffchainWorkerApi_offchain_worker
func OffchainWorkerApiOffchainWorker(dataPtr int32, dataLen int32) int64 {
	runtimeApi().
		Module(offchain_worker.ApiModuleName).(offchain_worker.Module).
		OffchainWorker(dataPtr, dataLen)

	return 0
}

//go:export GenesisBuilder_create_default_config
func GenesisBuilderCreateDefaultConfig(_, _ int32) int64 {
	return runtimeApi().
		Module(genesisbuilder.ApiModuleName).(genesisbuilder.Module).
		CreateDefaultConfig()
}

//go:export GenesisBuilder_build_config
func GenesisBuilderBuildConfig(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(genesisbuilder.ApiModuleName).(genesisbuilder.Module).
		BuildConfig(dataPtr, dataLen)
}

//go:export CollectCollationInfo_collect_collation_info
func CollectCollationInfoCollectCollationInfo(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(collect_collation_info.ApiModuleName).(collect_collation_info.Module).
		CollectCollationInfo(dataPtr, dataLen)
}

//go:export validate_block
func ParachainValidateBlock(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(parachain.ApiModuleName).(parachain.Module).
		ValidateBlock(dataPtr, dataLen)
}

//go:export Benchmark_dispatch
func BenchmarkDispatch(dataPtr int32, dataLen int32) int64 {
	return benchmarking.New(
		SystemIndex,
		modules,
		decoder,
		logger,
	).BenchmarkDispatch(dataPtr, dataLen)
}

//go:export Benchmark_hook
func BenchmarkHook(dataPtr int32, dataLen int32) int64 {
	return benchmarking.New(
		SystemIndex,
		modules,
		decoder,
		logger,
	).BenchmarkHook(dataPtr, dataLen)
}
