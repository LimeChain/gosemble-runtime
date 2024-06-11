package collect_collation_info

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/parachain_system"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "CollectCollationInfo"
	apiVersion    = 2
)

// Module implements the CollectCollationInfo Runtime API definition.
//
// For more information about API definition, see:
// https://github.com/paritytech/polkadot-sdk/blob/master/cumulus/primitives/core/src/lib.rs#L378
type Module struct {
	parachainSystem parachain_system.Module
	memUtils        utils.WasmMemoryTranslator
	logger          log.RuntimeLogger
}

func New(parachainSystem parachain_system.Module, logger log.RuntimeLogger) Module {
	return Module{
		memUtils:        utils.NewMemoryTranslator(),
		parachainSystem: parachainSystem,
		logger:          logger,
	}
}

// Name returns the name of the api module
func (m Module) Name() string {
	return ApiModuleName
}

// Item returns the first 8 bytes of the Blake2b hash of the name and version of the api module.
func (m Module) Item() primitives.ApiItem {
	hash := hashing.MustBlake2b8([]byte(ApiModuleName))
	return primitives.NewApiItem(hash, apiVersion)
}

// CollectCollationInfo returns the collation information by the header of a built block.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded block header.
// Returns a pointer-size of the SCALE-encoded collation information.
func (m Module) CollectCollationInfo(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	header, err := primitives.DecodeHeader(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	collationInfo, err := m.parachainSystem.CollectCollationInfo(header)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(collationInfo.Bytes())
}

func (m Module) Metadata() primitives.RuntimeApiMetadata {
	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name: "collect_collation_info",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "header",
					Type: sc.ToCompact(metadata.Header),
				},
			},
			Output: sc.ToCompact(metadata.TypesParachainValidationResult),
			Docs:   sc.Sequence[sc.Str]{"Returns the collation information by the block header."},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{" The collect collation info api."},
	}
}
