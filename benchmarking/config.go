package benchmarking

import "flag"

// cmd flags and other options related to benchmarking
var Config = initBenchmarkingConfig()

type benchmarkingConfig struct {
	Steps, Repeat, HeapPages, DbCache      int
	WasmRuntime, GC, TinyGoVersion, Target string
	GenerateWeightFiles                    bool
	Overhead                               overheadConfig
}

func initBenchmarkingConfig() benchmarkingConfig {
	cfg := benchmarkingConfig{}
	cfg.WasmRuntime = "../build/runtime.wasm"
	flag.IntVar(&cfg.Steps, "steps", 2, "Select how many samples we should take across the variable components.")
	flag.IntVar(&cfg.Repeat, "repeat", 1, "Select how many repetitions of this benchmark should run from within the wasm.")
	flag.IntVar(&cfg.HeapPages, "heap-pages", 4096, "Cache heap allocation pages.")
	flag.IntVar(&cfg.DbCache, "db-cache", 1024, "Limit the memory the database cache can use.")
	flag.StringVar(&cfg.GC, "gc", "", "GC flag used for building the runtime.")
	flag.StringVar(&cfg.TinyGoVersion, "tinygoversion", "", "TinyGO version used for building the runtime.")
	flag.StringVar(&cfg.Target, "target", "", "Target used for building the runtime.")
	flag.BoolVar(&cfg.GenerateWeightFiles, "generate-weight-files", true, "Whether to generate weight files.")
	cfg.Overhead = initOverheadConfig()
	return cfg
}

type overheadConfig struct {
	Warmup         int
	Repeat         int
	MaxExtPerBlock int
}

func initOverheadConfig() overheadConfig {
	cfg := overheadConfig{}
	flag.IntVar(&cfg.Warmup, "overhead.warmup", 10, "How many warmup rounds before measuring.")
	flag.IntVar(&cfg.Repeat, "overhead.repeat", 100, "How many times the benchmark test should be repeated.")
	flag.IntVar(&cfg.MaxExtPerBlock, "overhead.maxExtPerBlock", 200, "Maximum number of extrinsics per block")
	return cfg
}
