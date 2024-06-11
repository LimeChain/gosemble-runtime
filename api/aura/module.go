package aura

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/aura"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "AuraApi"
	apiVersion    = 1
)

// Module implements the AuraApi Runtime API definition.
//
// For more information about API definition, see:
// https://github.com/paritytech/polkadot-sdk/blob/master/substrate/primitives/consensus/aura/src/lib.rs#L86
type Module struct {
	aura     aura.AuraModule
	memUtils utils.WasmMemoryTranslator
	logger   log.RuntimeLogger
}

func New(aura aura.AuraModule, logger log.RuntimeLogger) Module {
	return Module{
		aura:     aura,
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

// Authorities returns current set of AuRa (Authority Round) authorities.
// Returns a pointer-size of the SCALE-encoded set of authorities.
func (m Module) Authorities() int64 {
	authorities, err := m.aura.StorageAuthoritiesBytes()
	if err != nil {
		m.logger.Critical(err.Error())
	}

	if !authorities.HasValue {
		return m.memUtils.BytesToOffsetAndSize([]byte{0})
	}

	return m.memUtils.BytesToOffsetAndSize(sc.SequenceU8ToBytes(authorities.Value))
}

// SlotDuration returns the slot duration for AuRa.
// Returns a pointer-size of the SCALE-encoded slot duration
func (m Module) SlotDuration() int64 {
	slotDuration := m.aura.SlotDuration()
	return m.memUtils.BytesToOffsetAndSize(slotDuration.Bytes())
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name:   "authorities",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.TypesSequenceU8),
			Docs:   sc.Sequence[sc.Str]{""},
		},
		primitives.RuntimeApiMethodMetadata{
			Name:   "slot_duration",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.PrimitiveTypesU64),
			Docs:   sc.Sequence[sc.Str]{},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{" The API to query account nonce."},
	}
}
