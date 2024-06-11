package core

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/executive"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "Core"
	apiVersion    = 4
)

type Core interface {
	Version() int64
	ExecuteBlock(dataPtr int32, dataLen int32)
	InitializeBlock(dataPtr int32, dataLen int32)
}

// Module implements the Core Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-runtime-core-module
type Module struct {
	executive      executive.Module
	decoder        types.RuntimeDecoder
	runtimeVersion *primitives.RuntimeVersion
	memUtils       utils.WasmMemoryTranslator
	mdGenerator    *primitives.MetadataTypeGenerator
	logger         log.RuntimeLogger
}

func New(module executive.Module, decoder types.RuntimeDecoder, runtimeVersion *primitives.RuntimeVersion, mdGenerator *primitives.MetadataTypeGenerator, logger log.RuntimeLogger) Module {
	return Module{
		executive:      module,
		decoder:        decoder,
		runtimeVersion: runtimeVersion,
		memUtils:       utils.NewMemoryTranslator(),
		mdGenerator:    mdGenerator,
		logger:         logger,
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

// Version returns a pointer-size SCALE-encoded Runtime version.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#defn-rt-core-version
func (m Module) Version() int64 {
	encoded := m.runtimeVersion.Bytes()

	return m.memUtils.BytesToOffsetAndSize(encoded)
}

// InitializeBlock starts the execution of a particular block.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded header of the block.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-core-initialize-block
func (m Module) InitializeBlock(dataPtr int32, dataLen int32) {
	data := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(data)
	header, err := primitives.DecodeHeader(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	err = m.executive.InitializeBlock(header)
	if err != nil {
		m.logger.Critical(err.Error())
	}
}

// ExecuteBlock executes the provided block.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded block.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-core-execute-block
func (m Module) ExecuteBlock(dataPtr int32, dataLen int32) {
	data := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(data)
	block, err := m.decoder.DecodeBlock(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	err = m.executive.ExecuteBlock(block)
	if err != nil {
		m.logger.Critical(err.Error())
	}
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	blockId, _ := m.mdGenerator.GetId("block")

	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name:   "version",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.TypesRuntimeVersion),
			Docs:   sc.Sequence[sc.Str]{" Returns the version of the runtime."},
		},
		primitives.RuntimeApiMethodMetadata{
			Name: "execute_block",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "block",
					Type: sc.ToCompact(blockId),
				},
			},
			Output: sc.ToCompact(metadata.TypesEmptyTuple),
			Docs:   sc.Sequence[sc.Str]{" Execute the given block."},
		},
		primitives.RuntimeApiMethodMetadata{
			Name: "initialize_block",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "header",
					Type: sc.ToCompact(metadata.Header),
				},
			},
			Output: sc.ToCompact(metadata.TypesEmptyTuple),
			Docs:   sc.Sequence[sc.Str]{" Initialize a block with the given header."},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{" The `Core` runtime api that every Substrate runtime needs to implement."},
	}
}
