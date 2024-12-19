module github.com/LimeChain/gosemble

go 1.21

require (
	github.com/ChainSafe/gossamer v0.9.0
	github.com/LimeChain/go-schnorrkel v1.1.0
	github.com/LimeChain/goscale v0.0.0-20230105112432-c7d2229e9977
	github.com/centrifuge/go-substrate-rpc-client/v4 v4.2.1
	github.com/gtank/merlin v0.1.1
	github.com/iancoleman/strcase v0.3.0
	github.com/montanaflynn/stats v0.7.1
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/stretchr/testify v1.9.0
	github.com/vedhavyas/go-subkey v1.0.4
	golang.org/x/crypto v0.28.0
)

require (
	filippo.io/edwards25519 v1.0.0-rc.1 // indirect
	github.com/ChainSafe/go-schnorrkel v1.1.0 // indirect
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/adrg/xdg v0.4.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.1 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.2 // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.0 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/base58 v1.0.5 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/ethereum/go-ethereum v1.14.3 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/getsentry/sentry-go v0.18.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/gtank/ristretto255 v0.1.3-0.20210930101514-6bb39798585c // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20220103164710-9a04d6ca976b // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.12.4 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/pierrec/xxHash v0.1.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/qdm12/gotree v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tetratelabs/wazero v1.1.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/vedhavyas/go-subkey/v2 v2.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.2.1 // indirect
)

replace github.com/ChainSafe/gossamer => ./gossamer

replace github.com/LimeChain/goscale => ./goscale

replace github.com/tetratelabs/wazero => github.com/ChainSafe/wazero v0.0.0-20240319130522-78b21a59bd5f

replace github.com/LimeChain/go-schnorrkel => ./go-schnorrkel
