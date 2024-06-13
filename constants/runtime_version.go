package constants

import (
	sc "github.com/LimeChain/goscale"
)

// If the runtime behavior changes, increment spec_version and set impl_version to 0.
// If only runtime implementation changes and behavior does not,
// then leave spec_version as is and increment impl_version.

const SpecName = "node-template"
const ImplName = "node-template"
const AuthoringVersion = 1
const SpecVersion = 100
const ImplVersion = 1
const TransactionVersion = 1
const StateVersion = 1
const StorageVersion = 0

const BlockHashCount = sc.U64(2400)
