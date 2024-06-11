package offchain_worker

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/executive"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "OffchainWorkerApi"
	apiVersion    = 2
)

// Module implements the OffchainWorkerApi Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-runtime-offchainapi-module
type Module struct {
	executive executive.Module
	memUtils  utils.WasmMemoryTranslator
	logger    log.RuntimeLogger
}

func New(executive executive.Module, logger log.RuntimeLogger) Module {
	return Module{
		executive: executive,
		memUtils:  utils.NewMemoryTranslator(),
		logger:    logger,
	}
}

// Name returns the name of the api module.
func (m Module) Name() string {
	return ApiModuleName
}

// Item returns the first 8 bytes of the Blake2b hash of the name and version of the api module.
func (m Module) Item() types.ApiItem {
	hash := hashing.MustBlake2b8([]byte(ApiModuleName))
	return types.NewApiItem(hash, apiVersion)
}

// OffchainWorker starts an off-chain task for an imported block.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded header of the block.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#id-offchainworkerapi_offchain_worker
func (m Module) OffchainWorker(dataPtr int32, dataLen int32) {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)
	header, err := primitives.DecodeHeader(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	err = m.executive.OffchainWorker(header)
	if err != nil {
		m.logger.Critical(err.Error())
	}
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name: "offchain_worker",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "header",
					Type: sc.ToCompact(metadata.Header),
				},
			},
			Output: sc.ToCompact(metadata.TypesEmptyTuple),
			Docs:   sc.Sequence[sc.Str]{" Starts the off-chain task for given block header."},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{" The offchain worker api."},
	}
}
