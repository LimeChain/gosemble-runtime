package constants

import (
	sc "github.com/LimeChain/goscale"
)

// If the runtime behavior changes, increment spec_version and set impl_version to 0.
// If only runtime implementation changes and behavior does not,
// then leave spec_version as is and increment impl_version.

const SpecName = "test-parachain"
const ImplName = "test-parachain"
const AuthoringVersion = 1
const SpecVersion = 1_007_000
const ImplVersion = 0
const TransactionVersion = 6
const StateVersion = 0
const StorageVersion = 0

const BlockHashCount = sc.U64(2400)
