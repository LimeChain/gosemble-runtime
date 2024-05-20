package main

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/api/account_nonce"
	apiBabe "github.com/LimeChain/gosemble/api/babe"
	"github.com/LimeChain/gosemble/api/benchmarking"
	blockbuilder "github.com/LimeChain/gosemble/api/block_builder"
	"github.com/LimeChain/gosemble/api/core"
	genesisbuilder "github.com/LimeChain/gosemble/api/genesis_builder"
	apiGrandpa "github.com/LimeChain/gosemble/api/grandpa"
	"github.com/LimeChain/gosemble/api/metadata"
	"github.com/LimeChain/gosemble/api/offchain_worker"
	"github.com/LimeChain/gosemble/api/session_keys"
	taggedtransactionqueue "github.com/LimeChain/gosemble/api/tagged_transaction_queue"
	apiTxPayments "github.com/LimeChain/gosemble/api/transaction_payment"
	apiTxPaymentsCall "github.com/LimeChain/gosemble/api/transaction_payment_call"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/execution/extrinsic"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/authorship"
	babe "github.com/LimeChain/gosemble/frame/babe"
	"github.com/LimeChain/gosemble/frame/balances"
	"github.com/LimeChain/gosemble/frame/executive"
	"github.com/LimeChain/gosemble/frame/grandpa"
	"github.com/LimeChain/gosemble/frame/session"
	session_historical "github.com/LimeChain/gosemble/frame/session_historical"
	"github.com/LimeChain/gosemble/frame/sudo"
	"github.com/LimeChain/gosemble/frame/system"
	sysExtensions "github.com/LimeChain/gosemble/frame/system/extensions"
	tm "github.com/LimeChain/gosemble/frame/testable"
	"github.com/LimeChain/gosemble/frame/timestamp"
	"github.com/LimeChain/gosemble/frame/transaction_payment"
	txExtensions "github.com/LimeChain/gosemble/frame/transaction_payment/extensions"
	"github.com/LimeChain/gosemble/hooks"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/log"
	sessiontypes "github.com/LimeChain/gosemble/primitives/session"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	BondingDuration               = 24 * 28
	SessionsPerEra                = 6
	BabeMaxAuthorities     sc.U32 = 100
	GrandpaMaxAuthorities         = 100
	GrandpaMaxNominators          = 64
	MaxSetIdSessionEntries        = BondingDuration * SessionsPerEra
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

const (
	SystemIndex sc.U8 = iota
	TimestampIndex
	SessionIndex
	BabeIndex
	GrandpaIndex
	BalancesIndex
	TxPaymentsIndex
	SudoIndex
	SessionHistoricalIndex
	AuthorshipIndex
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
	// The BABE epoch configuration at genesis.
	BabeGenesisEpochConfig = babetypes.EpochConfiguration{
		C:            constants.PrimaryProbability,
		AllowedSlots: babetypes.NewPrimarySlots(),
	}

	EpochDuration = constants.EpochDurationInSlots
)

var (
	logger      = log.NewLogger()
	mdGenerator = primitives.NewMetadataTypeGenerator()
	// Modules contains all the modules used by the runtime.
	modules = initializeModules()
	extra   = newSignedExtra()
	decoder = types.NewRuntimeDecoder(modules, extra, SudoIndex, logger)
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

// Construct runtime modules

func initializeModules() []primitives.Module {
	systemModule := system.New(
		SystemIndex,
		system.NewConfig(
			primitives.BlockHashCount{U32: sc.U32(constants.BlockHashCount)},
			blockWeights,
			blockLength,
			DbWeight,
			RuntimeVersion,
			maxConsumers,
		),
		mdGenerator,
		logger,
	)

	handler := session.NewHandler([]sessiontypes.OneSessionHandler{})

	periodicSession := session.NewPeriodicSessions(Period, Offset)

	sessionModule := session.New(
		SessionIndex,
		session.NewConfig(
			DbWeight,
			blockWeights,
			systemModule,
			periodicSession,
			handler,
			session.DefaultManager{},
		),
		mdGenerator,
		logger,
	)

	externalTrugger := babe.ExternalTrigger{}

	babeModule := babe.New(
		BabeIndex,
		babe.NewConfig(
			primitives.PublicKeySr25519,
			BabeGenesisEpochConfig,
			EpochDuration,
			externalTrugger,
			sessionModule,
			BabeMaxAuthorities,
			TimestampMinimumPeriod,
			systemModule.StorageDigest,
			systemModule,
		),
		mdGenerator,
		logger,
	)
	sessionModule.AppendHandlers(babeModule)

	sessionHistoricalModule := session_historical.New(
		SessionHistoricalIndex,
		session_historical.NewConfig(sessionModule),
		mdGenerator,
		logger,
	)

	sessionFindAccount := session.NewFindAccountFromAuthorIndex(sessionModule, babeModule)

	authorshipModule := authorship.New(
		AuthorshipIndex,
		authorship.NewConfig(
			sessionFindAccount,
			authorship.DefaulthEventHandler{}, // TODO: implemented by "imonline" module
			systemModule,
		),
		mdGenerator,
		logger,
	)

	grandpaEquivocationReportSystem := grandpa.NewEquivocationReportSystem(sessionHistoricalModule, authorshipModule, logger)

	grandpaModule := grandpa.New(
		GrandpaIndex,
		grandpa.NewConfig(
			primitives.PublicKeyEd25519,
			GrandpaMaxAuthorities,
			GrandpaMaxNominators,
			MaxSetIdSessionEntries,
			sessionHistoricalModule,
			grandpaEquivocationReportSystem,
			systemModule,
			sessionModule,
		),
		logger,
		mdGenerator,
	)
	grandpaEquivocationReportSystem.SetModule(grandpaModule)

	timestampModule := timestamp.New(
		TimestampIndex,
		timestamp.NewConfig(babeModule, DbWeight, TimestampMinimumPeriod),
		mdGenerator,
	)

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

	sudoModule := sudo.New(SudoIndex, sudo.NewConfig(DbWeight, systemModule), mdGenerator, logger)

	testableModule := tm.New(TestableIndex, mdGenerator)

	return []primitives.Module{
		systemModule,
		timestampModule,
		sessionModule,
		babeModule,
		grandpaModule,
		balancesModule,
		tpmModule,
		sudoModule,
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
	babeModule := primitives.MustGetModule(BabeIndex, modules).(babe.Module)
	grandpaModule := primitives.MustGetModule(GrandpaIndex, modules).(grandpa.Module)
	txPaymentsModule := primitives.MustGetModule(TxPaymentsIndex, modules).(transaction_payment.Module)

	executiveModule := executive.New(
		systemModule,
		runtimeExtrinsic,
		hooks.DefaultOnRuntimeUpgrade{},
		logger,
	)

	sessions := []primitives.Session{
		babeModule,
		grandpaModule,
	}

	coreApi := core.New(executiveModule, decoder, RuntimeVersion, mdGenerator, logger)
	blockBuilderApi := blockbuilder.New(runtimeExtrinsic, executiveModule, decoder, mdGenerator, logger)
	taggedTxQueueApi := taggedtransactionqueue.New(executiveModule, decoder, mdGenerator, logger)
	babeApi := apiBabe.New(babeModule, logger)
	grandpaApi := apiGrandpa.New(grandpaModule, logger)
	accountNonceApi := account_nonce.New(systemModule, logger)
	txPaymentsApi := apiTxPayments.New(decoder, txPaymentsModule, logger)
	txPaymentsCallApi := apiTxPaymentsCall.New(decoder, txPaymentsModule, logger)
	sessionKeysApi := session_keys.New(sessions, logger)
	offchainWorkerApi := offchain_worker.New(executiveModule, logger)

	genesisBuilderApi := genesisbuilder.New(modules, logger)

	metadataApi := metadata.New(
		runtimeExtrinsic,
		[]primitives.RuntimeApiModule{
			coreApi,
			blockBuilderApi,
			taggedTxQueueApi,
			babeApi,
			grandpaApi,
			accountNonceApi,
			txPaymentsApi,
			txPaymentsCallApi,
			sessionKeysApi,
			offchainWorkerApi,
		},
		logger,
		mdGenerator,
	)

	apis := []primitives.ApiModule{
		coreApi,
		blockBuilderApi,
		taggedTxQueueApi,
		metadataApi,
		babeApi,
		grandpaApi,
		accountNonceApi,
		txPaymentsApi,
		txPaymentsCallApi,
		sessionKeysApi,
		offchainWorkerApi,
		genesisBuilderApi,
	}

	runtimeApi := types.NewRuntimeApi(apis, logger)

	RuntimeVersion.SetApis(runtimeApi.Items())

	return runtimeApi
}

// Implement runtime APIs

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

//go:export BabeApi_configuration
func BabeApiConfiguration(_, _ int32) int64 {
	return runtimeApi().
		Module(apiBabe.ApiModuleName).(apiBabe.Module).
		Configuration()
}

//go:export BabeApi_current_epoch_start
func BabeApiCurrentEpochStart(_, _ int32) int64 {
	return runtimeApi().
		Module(apiBabe.ApiModuleName).(apiBabe.Module).
		CurrentEpochStart()
}

//go:export BabeApi_current_epoch
func BabeApiCurrentEpoch(_, _ int32) int64 {
	return runtimeApi().
		Module(apiBabe.ApiModuleName).(apiBabe.Module).
		CurrentEpoch()
}

//go:export BabeApi_next_epoch
func BabeApiNextEpoch(_, _ int32) int64 {
	return runtimeApi().
		Module(apiBabe.ApiModuleName).(apiBabe.Module).
		NextEpoch()
}

// TODO: implement
// //go:export BabeApi_generate_key_ownership_proof
// func BabeApiGenerateKeyOwnershipProof(dataPtr int32, dataLen int32) int64 {
// 	return runtimeApi().
// 		Module(apiBabe.ApiModuleName).(apiBabe.Module).
// 		GenerateKeyOwnershipProof(dataPtr, dataLen)
// }

// //go:export BabeApi_submit_report_equivocation_unsigned_extrinsic
// func BabeApiSubmitReportEquivocationUnsignedExtrinsic(dataPtr int32, dataLen int32) int64 {
// 	return runtimeApi().
// 		Module(apiBabe.ApiModuleName).(apiBabe.Module).
// 		SubmitReportEquivocationUnsignedExtrinsic(dataPtr, dataLen)
// }

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

//go:export GrandpaApi_current_set_id
func GrandpaApiCurrentSetId() int64 {
	return runtimeApi().
		Module(apiGrandpa.ApiModuleName).(apiGrandpa.Module).
		CurrentSetId()
}

//go:export GrandpaApi_submit_report_equivocation_unsigned_extrinsic
func GrandpaApi_submit_report_equivocation_unsigned_extrinsic(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(apiGrandpa.ApiModuleName).(apiGrandpa.Module).
		SubmitReportEquivocationUnsignedExtrinsic(dataPtr, dataLen)
}

//go:export GrandpaApi_generate_key_ownership_proof
func GrandpaApiGenerateKeyOwnershipProof(dataPtr int32, dataLen int32) int64 {
	return runtimeApi().
		Module(apiGrandpa.ApiModuleName).(apiGrandpa.Module).
		GenerateKeyOwnershipProof(dataPtr, dataLen)
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
