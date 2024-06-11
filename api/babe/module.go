package babe

import (
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/babe"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "BabeApi"
	apiVersion    = 2
)

// Module implements the BabeApi Runtime API definition, necessary for block authorship.
type Module struct {
	babe     babe.Module
	memUtils utils.WasmMemoryTranslator
	logger   log.RuntimeLogger
}

func New(babe babe.Module, logger log.RuntimeLogger) Module {
	return Module{
		babe:     babe,
		memUtils: utils.NewMemoryTranslator(),
		logger:   logger,
	}
}

// Name returns the name of the api module.
func (m Module) Name() string {
	return ApiModuleName
}

// Item returns the first 8 bytes of the Blake2b hash of the name and version of the api module.
func (m Module) Item() primitives.ApiItem {
	hash := hashing.MustBlake2b8([]byte(ApiModuleName))
	return primitives.NewApiItem(hash, apiVersion)
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name:   "configuration",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.TypesBabeConfiguration),
			Docs:   sc.Sequence[sc.Str]{""},
		},
		primitives.RuntimeApiMethodMetadata{
			Name:   "current_epoch_start",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.PrimitiveTypesU64),
			Docs:   sc.Sequence[sc.Str]{""},
		},
		primitives.RuntimeApiMethodMetadata{
			Name:   "current_epoch",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.TypesBabeEpoch),
			Docs:   sc.Sequence[sc.Str]{""},
		},
		primitives.RuntimeApiMethodMetadata{
			Name:   "next_epoch",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.TypesBabeEpoch),
			Docs:   sc.Sequence[sc.Str]{""},
		},
		// TODO: add metadata for GenerateKeyOwnershipProof and SubmitReportEquivocationUnsignedExtrinsic
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{"Babe consensus API module."},
	}
}

// Returns configuration data used by the Babe consensus engine.
func (m Module) Configuration() int64 {
	epochConfig, err := m.babe.StorageEpochConfig()
	if err != nil {
		m.logger.Critical(err.Error())
	}

	if reflect.DeepEqual(epochConfig, babetypes.EpochConfiguration{}) {
		epochConfig = m.babe.EpochConfig()
	}

	authorities, err := m.babe.StorageAuthorities()
	if err != nil {
		m.logger.Critical(err.Error())
	}

	randomness, err := m.babe.StorageRandomness()
	if err != nil {
		m.logger.Critical(err.Error())
	}

	config := babetypes.Configuration{
		// The slot duration in milliseconds. Currently, only the value provided by this
		// type at genesis will be used. Dynamic slot duration may be supported in the future.
		SlotDuration: m.babe.SlotDuration(),
		// The duration of epochs in slots.
		EpochLength: m.babe.EpochDuration(),
		// A constant value that is used in the threshold calculation formula as defined in
		// https://spec.polkadot.network/sect-block-production#defn-babe-constant
		C: epochConfig.C,
		// The authority list for the genesis epoch as defined in
		// https://spec.polkadot.network/chap-sync#defn-authority-list
		Authorities: authorities,
		// The randomness for the genesis epoch.
		Randomness: randomness,
		// Whether this chain should run with a round-robin-style secondary slot and
		// if this secondary slot requires the inclusion of an auxiliary VRF output.
		AllowedSlots: epochConfig.AllowedSlots,
	}

	return m.memUtils.BytesToOffsetAndSize(config.Bytes())
}

// Returns the start slot of the current epoch.
func (m Module) CurrentEpochStart() int64 {
	epochStart, err := m.babe.CurrentEpochStart()
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(epochStart.Bytes())
}

// Returns information regarding the current epoch.
func (m Module) CurrentEpoch() int64 {
	epoch, err := m.babe.CurrentEpoch()
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(epoch.Bytes())
}

// Returns information regarding the next epoch (which was already
// previously announced).
func (m Module) NextEpoch() int64 {
	epoch, err := m.babe.NextEpoch()
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(epoch.Bytes())
}

// Generates a proof of key ownership for the given authority in the
// current epoch. An example usage of this module is coupled with the
// session historical module to prove that a given authority key is
// tied to a given staking identity during a specific session. Proofs
// of key ownership are necessary for submitting equivocation reports.
// NOTE: even though the API takes a `slot` as parameter the current
// implementations ignores this parameter and instead relies on this
// method being called at the correct block height, i.e. any point at
// which the epoch for the given slot is live on-chain. Future
// implementations will instead use indexed data through an offchain
// worker, not requiring older states to be available.
func (m Module) GenerateKeyOwnershipProof(dataPtr int32, dataLen int32) int64 {
	// TODO: Implement

	m.logger.Critical("GenerateKeyOwnershipProof is not implemented")

	return m.memUtils.BytesToOffsetAndSize(sc.NewOption[sc.Empty](nil).Bytes())
}

// Submits an unsigned extrinsic to report an equivocation. The caller
// must provide the equivocation proof and a key ownership proof
// (should be obtained using `generate_key_ownership_proof`). The
// extrinsic will be unsigned and should only be accepted for local
// authorship (not to be broadcast to the network). This method returns
// `None` when creation of the extrinsic fails, e.g. if equivocation
// reporting is disabled for the given runtime (i.e. this method is
// hardcoded to return `None`). Only useful in an offchain context.
func (m Module) SubmitReportEquivocationUnsignedExtrinsic(dataPtr int32, dataLen int32) int64 {
	// TODO: Implement

	m.logger.Critical("SubmitReportEquivocationUnsignedExtrinsic is not implemented")

	return m.memUtils.BytesToOffsetAndSize(sc.NewOption[sc.Empty](nil).Bytes())
}
