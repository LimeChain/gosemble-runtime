package benchmarking

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/benchmarking"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
	"github.com/montanaflynn/stats"
)

// Module implements API for benchmarking dispatchable extrinsic calls and system hooks.
// For more information about the benchmarking process, see:
// /docs/docs/development/benchmarking.md
type Module struct {
	modules       []primitives.Module
	systemModule  system.Module
	transactional support.Transactional[primitives.PostDispatchInfo]
	decoder       types.RuntimeDecoder
	memUtils      utils.WasmMemoryTranslator
	hashing       io.Hashing
	logger        log.RuntimeLogger
}

func New(systemIndex sc.U8, modules []primitives.Module, decoder types.RuntimeDecoder, storage io.Storage, transactionBroker io.TransactionBroker, logger log.RuntimeLogger) Module {
	systemModule := primitives.MustGetModule(systemIndex, modules).(system.Module)

	return Module{
		modules:       modules,
		systemModule:  systemModule,
		decoder:       decoder,
		transactional: support.NewTransactional[primitives.PostDispatchInfo](storage, transactionBroker, logger),
		memUtils:      utils.NewMemoryTranslator(),
		hashing:       io.NewHashing(),
		logger:        logger,
	}
}

// BenchmarkDispatch benchmarks the execution of dispatchable calls.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded benchmarking configuration.
//
// Executes a dispatch extrinsic call N times in isolated environment
// by measuring the elapsed time in each execution.
//
// Returns a pointer-size of the SCALE-encoded benchmarking result.
//
// TODO:
// Implement DbCommit, DbWipe once the state implementation
// in Gossamer supports caching and nested transactions.
// https://github.com/ChainSafe/gossamer/discussions/3646
func (m Module) BenchmarkDispatch(dataPtr int32, dataLen int32) int64 {
	data := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(data)

	config, err := benchmarking.DecodeBenchmarkConfig(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	buffer = bytes.NewBuffer(sc.SequenceU8ToBytes(config.Benchmark))
	extrinsic, err := m.decoder.DecodeUncheckedExtrinsic(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	function := extrinsic.Function()
	args := function.Args()

	benchmarkResult := m.executeBenchmark(config, func(origin primitives.RawOrigin) float64 {
		var start, end int64

		start = benchmarking.CurrentTime()
		_, err := m.transactional.WithStorageLayer(
			func() (primitives.PostDispatchInfo, error) {
				return function.Dispatch(origin, args)
			},
		)
		end = benchmarking.CurrentTime()
		if err != nil {
			m.logger.Critical(err.Error())
		}

		return float64(end - start)
	})

	return m.memUtils.BytesToOffsetAndSize(benchmarkResult.Bytes())
}

// BenchmarkHook benchmarks the execution of system hooks.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded benchmarking configuration.
//
// Based on the provided system hook, it executes it for each of
// the preconfigured runtime modules and sums their execution elapsed time.
//
// Returns a pointer-size of the SCALE-encoded benchmarking result.
func (m Module) BenchmarkHook(dataPtr int32, dataLen int32) int64 {
	data := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(data)

	config, err := benchmarking.DecodeBenchmarkConfig(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	buffer = bytes.NewBuffer(sc.SequenceU8ToBytes(config.Benchmark))
	hook, err := sc.DecodeStr(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	arg0, err := sc.DecodeU64(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	arg1, err := primitives.DecodeWeight(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	benchmarkResult := m.executeBenchmark(config, func(origin primitives.RawOrigin) float64 {
		var elapsed float64

		// Benchmark the cumulative time of dispatchable module hooks.
		switch hook {
		case "on_initialize":
			elapsed = measureHooks(m.modules, func(module primitives.DispatchModule) error {
				_, err := module.OnInitialize(arg0)
				return err
			}, m.logger)
		case "on_runtime_upgrade":
			elapsed = measureHooks(m.modules, func(module primitives.DispatchModule) error {
				_ = module.OnRuntimeUpgrade()
				return nil
			}, m.logger)
		case "on_finalize":
			elapsed = measureHooks(m.modules, func(module primitives.DispatchModule) error {
				err := module.OnFinalize(arg0)
				return err
			}, m.logger)
		case "on_idle":
			elapsed = measureHooks(m.modules, func(module primitives.DispatchModule) error {
				_ = module.OnIdle(arg0, arg1)
				return nil
			}, m.logger)
		default:
			m.logger.Critical("unsupported hook")
		}

		return elapsed
	})

	return m.memUtils.BytesToOffsetAndSize(benchmarkResult.Bytes())
}

func measureHooks(modules []primitives.Module, hookFn func(module primitives.DispatchModule) error, logger log.RuntimeLogger) float64 {
	var start, end int64

	for _, module := range modules {
		if start == 0 || end == 0 {
			start = benchmarking.CurrentTime()
			end = start
		}

		t0 := benchmarking.CurrentTime()
		err := hookFn(module)
		t1 := benchmarking.CurrentTime()
		elapsed := t1 - t0
		end += elapsed

		if err != nil {
			logger.Critical(err.Error())
		}
	}

	return float64(end - start)
}

func (m Module) executeBenchmark(config benchmarking.BenchmarkConfig, fn func(origin primitives.RawOrigin) float64) benchmarking.BenchmarkResult {
	origin, accountId := m.originAndMaybeAccount(config)

	measuredDurations := []float64{}

	benchmarking.StoreSnapshotDb()

	// Always do at least one internal repeat.
	repeats := int(config.InternalRepeats)
	if repeats < 1 {
		repeats = 1
	}
	for i := 1; i <= repeats; i++ {
		// The dispatch call is executed in a transactional context,
		// allowing to rollback and reset the state after each iteration.
		// as an alternative of providing before hook.

		benchmarking.RestoreSnapshotDb()

		// Does nothing, for now
		benchmarking.WipeDb()

		// Set up the externalities environment for the setup we want to
		// benchmark.

		// Sets the block number to 1 to allow emitting events
		m.systemModule.StorageBlockNumberSet(1)

		// Commit the externalities to the database, flushing the DB cache.
		// This will enable worst case scenario for reading from the database.
		// Does nothing, for now
		benchmarking.CommitDb()

		// Whitelist known storage keys.
		m.whitelistWellKnownKeys()

		// Whitelist the signer account key.
		if accountId.HasValue {
			keyStorageAccount := m.accountStorageKeyFrom(accountId.Value)
			benchmarking.SetWhitelist(keyStorageAccount)
		}

		// Reset the read/write counter so we don't count
		// operations in the setup process.
		benchmarking.ResetReadWriteCount()

		benchmarking.StartDbTracker()

		elapsed := fn(origin)

		// Calculate the diff caused by the benchmark.
		measuredDurations = append(measuredDurations, elapsed)

		benchmarking.StopDbTracker()

		// Commit the changes to get proper write count.
		// Does nothing, for now
		benchmarking.CommitDb()
	}

	// Calculate the average time.
	time, err := stats.Mean(measuredDurations)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return benchmarking.BenchmarkResult{
		Time:   sc.NewU128(int64(time)),
		Reads:  sc.U32(benchmarking.DbReadCount()),
		Writes: sc.U32(benchmarking.DbWriteCount()),
	}
}

func (m Module) originAndMaybeAccount(benchmarkConfig benchmarking.BenchmarkConfig) (primitives.RawOrigin, sc.Option[primitives.AccountId]) {
	// TODO: pass the origin as an option
	origin := benchmarkConfig.Origin

	if origin.IsSignedOrigin() {
		id, err := origin.AsSigned()
		if err != nil {
			m.logger.Critical(err.Error())
		}
		return origin, sc.NewOption[primitives.AccountId](id)
	}

	return origin, sc.NewOption[primitives.AccountId](nil)
}

func (m Module) whitelistWellKnownKeys() {
	keySystemHash := m.hashing.Twox128([]byte("System"))
	keyBalancesHash := m.hashing.Twox128([]byte("Balances"))
	keyBlockWeight := m.hashing.Twox128([]byte("BlockWeight"))
	keyExecutionPhaseHash := m.hashing.Twox128([]byte("ExecutionPhase"))
	keyEventCountHash := m.hashing.Twox128([]byte("EventCount"))
	keyEventsHash := m.hashing.Twox128([]byte("Events"))
	keyNumberHash := m.hashing.Twox128([]byte("Number"))
	keyTotalIssuanceHash := m.hashing.Twox128([]byte("TotalIssuance"))

	benchmarking.SetWhitelist(append(keyBalancesHash, keyTotalIssuanceHash...))
	benchmarking.SetWhitelist(append(keySystemHash, keyBlockWeight...))
	benchmarking.SetWhitelist(append(keySystemHash, keyNumberHash...))
	benchmarking.SetWhitelist(append(keySystemHash, keyExecutionPhaseHash...))
	benchmarking.SetWhitelist(append(keySystemHash, keyEventCountHash...))
	benchmarking.SetWhitelist(append(keySystemHash, keyEventsHash...))

	benchmarking.SetWhitelist([]byte(":transaction_level:"))
	benchmarking.SetWhitelist([]byte(":extrinsic_index"))
	benchmarking.SetWhitelist([]byte(":intrablock_entropy"))
}

func (m Module) accountStorageKeyFrom(address primitives.AccountId) []byte {
	addressBytes := address.FixedSequence.Bytes()
	keySystemHash := m.hashing.Twox128([]byte("System"))
	keyAccountHash := m.hashing.Twox128([]byte("Account"))
	addressHash := m.hashing.Blake128(addressBytes)
	keyStorageAccount := append(keySystemHash, keyAccountHash...)
	keyStorageAccount = append(keyStorageAccount, addressHash...)
	keyStorageAccount = append(keyStorageAccount, addressBytes...)
	return keyStorageAccount
}
