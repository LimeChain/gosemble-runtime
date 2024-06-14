CURRENT_DIR = $(shell pwd)

# runtime template configuration
RUNTIME_TEMPLATE_DIR = runtime/templates/poa
BUILD_PATH = build
RUNTIME_WASM = runtime.wasm
RUNTIME_WASM_BENCHMARKS = runtime-benchmarks.wasm

# docker image configuration
SRC_DIR = /src/examples/wasm/gosemble
IMAGE = tinygo/${TARGET}

# tinygo compiler configuration
VERSION = 0.31.0-dev
TARGET = polkawasm
OPT_LEVEL = s # 0, 1, 2, s, z
GC = extalloc # extalloc, extalloc_leaking
WASMOPT_PATH = /tinygo/lib/binaryen/bin/wasm-opt

# runtime build commands
DOCKER_BUILD_TINYGO = docker build --tag $(IMAGE):$(VERSION)-$(GC) -f tinygo/Dockerfile.$(TARGET) tinygo
DOCKER_RUN_TINYGO = docker run --rm -v $(CURRENT_DIR):$(SRC_DIR) -w $(SRC_DIR) $(IMAGE):$(VERSION)-$(GC) /bin/bash -c

TINYGO_BUILD_COMMAND_NODEBUG = tinygo build --no-debug -opt=$(OPT_LEVEL) -gc=$(GC) -target=$(TARGET)
TINYGO_BUILD_COMMAND = tinygo build -opt=$(OPT_LEVEL) -gc=$(GC) -target=$(TARGET)

RUNTIME_BUILD_NODEBUG = "WASMOPT="$(WASMOPT_PATH)" $(TINYGO_BUILD_COMMAND_NODEBUG) -o=$(SRC_DIR)/$(BUILD_PATH)/$(RUNTIME_WASM) $(SRC_DIR)/runtime/"
RUNTIME_BUILD = "WASMOPT="$(WASMOPT_PATH)" $(TINYGO_BUILD_COMMAND) -o=$(SRC_DIR)/$(BUILD_PATH)/$(RUNTIME_WASM) $(SRC_DIR)/runtime/"
RUNTIME_BUILD_BENCHMARKING = "WASMOPT="$(WASMOPT_PATH)" $(TINYGO_BUILD_COMMAND_NODEBUG) -tags=benchmarking -o=$(SRC_DIR)/$(BUILD_PATH)/$(RUNTIME_WASM_BENCHMARKS) $(SRC_DIR)/runtime/"

clear-wasi-libc:
	@cd tinygo/lib/wasi-libc && \
	make clean

clear-binaryen:
	@cd tinygo/lib/binaryen && \
	rm -rf CMakeCache.txt

build-docker-release: clear-binaryen
	@set -e; \
	$(DOCKER_BUILD_TINYGO);
	$(DOCKER_RUN_TINYGO) $(RUNTIME_BUILD_NODEBUG); \
	echo "Build - tinygo version: ${VERSION}, gc: ${GC} (no debug)"
	
build-docker-dev: clear-binaryen
	@set -e; \
	$(DOCKER_BUILD_TINYGO);
	$(DOCKER_RUN_TINYGO) $(RUNTIME_BUILD); \
	echo "Build - tinygo version: ${VERSION}, gc: ${GC}"
	
build-docker-benchmarking: clear-binaryen
	@set -e; \
	$(DOCKER_BUILD_TINYGO);
	$(DOCKER_RUN_TINYGO) $(RUNTIME_BUILD_BENCHMARKING); \
	echo "Build - tinygo version: ${VERSION}, gc: ${GC} (no debug) (benchmarking)"

build-wasi-libc: clear-wasi-libc
	@cd tinygo/lib/wasi-libc && \
	if [ ! -e Makefile ]; then \
		echo "Submodules have not been downloaded. Please download them using:\n git submodule update --init"; \
		exit 1; \
	fi && \
	echo "Building \"wasi-libc\""; \
	make -j4 EXTRA_CFLAGS="-O2 -g -DNDEBUG" MALLOC_IMPL=none; \

build-binaryen: clear-binaryen
	@cd tinygo/lib/binaryen && \
	if [ ! -e Makefile ]; then \
		echo "Submodules have not been downloaded. Please download them using:\n git submodule update --init"; \
		exit 1; \
	fi && \
	echo "Building \"binaryen\""; \
	cmake . && make; \

build-tinygo:
	@cd tinygo && \
	if [ ! -e lib/wasi-libc/sysroot ]; then \
		echo "Need to build wasi-libc. Please run: \"make build-wasi-libc\""; \
		exit 1; \
	fi; \
	if [ ! -e lib/binaryen/bin/wasm-opt ]; then \
		echo "Need to build binaryen. Please run: \"make build-binaryen\""; \
		exit 1; \
	fi; \
	echo "Building \"tinygo\""; \
	go install; \
	tinygo version; \

build-release: build-tinygo
	@echo "Building \"$(RUNTIME_WASM)\" (no-debug)"; \
	WASMOPT="$(CURRENT_DIR)/$(WASMOPT_PATH)" $(TINYGO_BUILD_COMMAND_NODEBUG) -o=$(BUILD_PATH)/$(RUNTIME_WASM) $(RUNTIME_TEMPLATE_DIR)/runtime.go

build-dev: build-tinygo
	@echo "Building \"$(RUNTIME_WASM)\""; \
	WASMOPT="$(CURRENT_DIR)/$(WASMOPT_PATH)" $(TINYGO_BUILD_COMMAND) -o=$(BUILD_PATH)/$(RUNTIME_WASM) $(RUNTIME_TEMPLATE_DIR)/runtime.go

build-benchmarking: build-tinygo
	@echo "Building \"$(RUNTIME_WASM_BENCHMARKS)\" (no-debug)"; \
	WASMOPT="$(CURRENT_DIR)/$(WASMOPT_PATH)" $(TINYGO_BUILD_COMMAND_NODEBUG) -tags benchmarking -o=$(BUILD_PATH)/$(RUNTIME_WASM_BENCHMARKS) $(RUNTIME_TEMPLATE_DIR)/runtime.go

test-coverage:
	@set -e; \
	./scripts/coverage.sh

test: test-unit test-integration

test-unit:
	@go test --tags "nonwasmenv" -cover -v `go list ./... | grep -v runtime`

test-integration:
	@go test --tags="nonwasmenv" -v -count=1 ./$(RUNTIME_TEMPLATE_DIR)/...

GENERATE_WEIGHT_FILES = true
benchmark: build-benchmarking
	@go test --tags="nonwasmenv" -bench=. ./$(RUNTIME_TEMPLATE_DIR)/... -run=XXX -benchtime=1x \
	-steps=50 \
	-repeat=20 \
	-heap-pages=4096 \
	-db-cache=1024 \
	-gc=$(GC) \
	-target=$(TARGET) \
	-tinygoversion=$(VERSION) \
	-generate-weight-files=$(GENERATE_WEIGHT_FILES);

benchmark-overhead: build-benchmarking
	@go test --tags="nonwasmenv" -bench=^BenchmarkOverhead ./benchmarking/... -run=^a -benchtime=1x \
	-gc=$(GC) \
	-target=$(TARGET) \
	-tinygoversion=$(VERSION) \
	-generate-weight-files=$(GENERATE_WEIGHT_FILES)

# substrate node configuration
SUBSTRATE_CHAIN_SPEC = local
substrate-build:
	cp $(BUILD_PATH)/$(RUNTIME_WASM) polkadot-sdk/substrate/bin/node-template/$(RUNTIME_WASM); \
	cd polkadot-sdk/substrate/bin/node-template/node; \
	cargo build --release

substrate-start-alice:
	cd polkadot-sdk; \
	./target/release/node-template purge-chain --base-path /tmp/alice --chain $(SUBSTRATE_CHAIN_SPEC) -y; \
	WASMTIME_BACKTRACE_DETAILS=1 RUST_LOG=runtime=trace ./target/release/node-template --execution=wasm \
	--base-path /tmp/alice \
	--chain $(SUBSTRATE_CHAIN_SPEC) \
	--alice \
	--port 30333 \
	--rpc-port 9945 \
	--validator

substrate-start-bob:
	cd polkadot-sdk; \
	./target/release/node-template purge-chain --base-path /tmp/bob --chain $(SUBSTRATE_CHAIN_SPEC) -y; \
	WASMTIME_BACKTRACE_DETAILS=1 RUST_LOG=runtime=trace ./target/release/node-template --execution=wasm \
	--base-path /tmp/bob \
	--chain $(SUBSTRATE_CHAIN_SPEC) \
	--bob \
	--port 30334 \
	--rpc-port 9946 \
	--validator

start-network-aura:
	cd ../../../..; \
	WASMTIME_BACKTRACE_DETAILS=1 RUST_LOG=runtime=trace ./target/release/node-template --dev --execution=wasm

start-network-babe:
	cd polkadot-sdk/substrate/bin/node/cli; \
	cargo build --release; \
	cd ../../../..; \
	WASMTIME_BACKTRACE_DETAILS=1 RUST_LOG=runtime=trace ./target/release/substrate-node --dev --execution=wasm

parachain-build:
	cp $(BUILD_PATH)/parachain.wasm polkadot-sdk/cumulus/polkadot-parachain/src/chain_spec/runtime.wasm; \
	cd polkadot-sdk; \
	cargo build --release -p polkadot-parachain-bin

# gossamer node configuration
CHAIN_SPEC_PLAIN = ../testdata/chain-spec/plain.json
CHAIN_SPEC_UPDATED = ../testdata/chain-spec/plain-updated.json
CHAIN_SPEC_RAW = ../testdata/chain-spec/raw-updated.json
GOSSAMER_BASE_PATH = tmp/gossamer

# use runtime build from the pos template
gossamer-build:
	@cd gossamer; \
	make build;

gossamer-import-runtime:
	@cd gossamer; \
	rm -f $(CHAIN_SPEC_UPDATED); \
	./bin/gossamer import-runtime --wasm-file ../$(BUILD_PATH)/$(RUNTIME_WASM) --chain $(CHAIN_SPEC_PLAIN) > $(CHAIN_SPEC_UPDATED); \
	rm -f $(CHAIN_SPEC_RAW); \
	./bin/gossamer build-spec --chain $(CHAIN_SPEC_UPDATED) --raw --output-path $(CHAIN_SPEC_RAW)

gossamer-init:
	@cd gossamer; \
	rm -rf $(GOSSAMER_BASE_PATH); \
	./bin/gossamer init --force \
	--base-path $(GOSSAMER_BASE_PATH) \
	--chain $(CHAIN_SPEC_RAW) \
	--key alice;

gossamer-start: gossamer-build gossamer-import-runtime gossamer-init
	@cd gossamer; \
	./bin/gossamer \
		--base-path $(GOSSAMER_BASE_PATH) \
		--rpc-external \
		--ws-external \
		--ws-port 8546 \
		--key alice;