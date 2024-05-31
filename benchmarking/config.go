package benchmarking

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

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

	buildDir, err := getBuildDir()
	if err != nil {
		log.Fatal(err)
	}
	cfg.WasmRuntime = filepath.Join(buildDir, "runtime.wasm")

	flag.IntVar(&cfg.Steps, "steps", 50, "Select how many samples we should take across the variable components.")
	flag.IntVar(&cfg.Repeat, "repeat", 20, "Select how many repetitions of this benchmark should run from within the wasm.")
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

func getBuildDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	for {
		// check if the directory exists
		buildDirPath := filepath.Join(currentDir, "build")
		stat, err := os.Stat(buildDirPath)
		if err == nil && stat.IsDir() {
			return buildDirPath, nil
		}

		// move up one level
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// reached the root directory
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("could not find \"build\" directory")
}
