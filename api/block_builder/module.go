package blockbuilder

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/execution/extrinsic"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/executive"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "BlockBuilder"
	apiVersion    = 6
)

type BlockBuilder interface {
	ApplyExtrinsic(dataPtr int32, dataLen int32) int64
	FinalizeBlock() int64
	InherentExtrinsics(dataPtr int32, dataLen int32) int64
	CheckInherents(dataPtr int32, dataLen int32) int64
}

// Module implements the BlockBuilder Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-runtime-blockbuilder-module
type Module struct {
	runtimeExtrinsic extrinsic.RuntimeExtrinsic
	executive        executive.Module
	decoder          types.RuntimeDecoder
	memUtils         utils.WasmMemoryTranslator
	mdGenerator      *primitives.MetadataTypeGenerator
	logger           log.RuntimeLogger
}

func New(runtimeExtrinsic extrinsic.RuntimeExtrinsic, executive executive.Module, decoder types.RuntimeDecoder, mdGenerator *primitives.MetadataTypeGenerator, logger log.RuntimeLogger) Module {
	return Module{
		runtimeExtrinsic: runtimeExtrinsic,
		executive:        executive,
		decoder:          decoder,
		memUtils:         utils.NewMemoryTranslator(),
		mdGenerator:      mdGenerator,
		logger:           logger,
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

// ApplyExtrinsic applies an extrinsic to a particular block.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded unchecked extrinsic.
// Returns a pointer-size of the SCALE-encoded result, which specifies if this extrinsic is included in this block or not.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-apply-extrinsic
func (m Module) ApplyExtrinsic(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	uxt, err := m.decoder.DecodeUncheckedExtrinsic(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	err = m.executive.ApplyExtrinsic(uxt)

	var applyExtrinsicResult primitives.ApplyExtrinsicResult
	switch typedErr := err.(type) {
	case nil:
		dispatchOutcome := primitives.DispatchOutcome(sc.NewVaryingData(sc.Empty{}))
		applyExtrinsicResult, err = primitives.NewApplyExtrinsicResult(dispatchOutcome)
	case primitives.TransactionValidityError:
		applyExtrinsicResult, err = primitives.NewApplyExtrinsicResult(typedErr)
	case primitives.DispatchError:
		dispatchOutcome := primitives.DispatchOutcome(sc.NewVaryingData(typedErr))
		applyExtrinsicResult, err = primitives.NewApplyExtrinsicResult(dispatchOutcome)
	}
	if err != nil {
		m.logger.Critical(err.Error())
	}

	buffer.Reset()
	err = applyExtrinsicResult.Encode(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(buffer.Bytes())
}

// FinalizeBlock finalizes the state changes for the current block.
// Returns a pointer-size of the SCALE-encoded header for this block.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#defn-rt-blockbuilder-finalize-block
func (m Module) FinalizeBlock() int64 {
	header, err := m.executive.FinalizeBlock()
	if err != nil {
		m.logger.Critical(err.Error())
	}
	encodedHeader := header.Bytes()
	return m.memUtils.BytesToOffsetAndSize(encodedHeader)
}

// InherentExtrinsics generates inherent extrinsics. Inherent data varies depending on chain configuration.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded inherent data.
// Returns a pointer-size of the SCALE-encoded timestamp extrinsic.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#defn-rt-builder-inherent-extrinsics
func (m Module) InherentExtrinsics(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	inherentData, err := primitives.DecodeInherentData(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	result, err := m.runtimeExtrinsic.CreateInherents(*inherentData)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(result)
}

// CheckInherents checks the inherents are valid.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded inherent data.
// Returns a pointer-size of the SCALE-encoded result, specifying if all inherents are valid.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#id-blockbuilder_check_inherents
func (m Module) CheckInherents(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	block, err := m.decoder.DecodeBlock(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	inherentData, err := primitives.DecodeInherentData(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	result, err := m.runtimeExtrinsic.CheckInherents(*inherentData, block)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(result.Bytes())
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	blockId, _ := m.mdGenerator.GetId("block")

	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name: "apply_extrinsic",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "Extrinsic",
					Type: sc.ToCompact(metadata.UncheckedExtrinsic),
				},
			},
			Output: sc.ToCompact(metadata.TypesApplyExtrinsicResult),
			Docs: sc.Sequence[sc.Str]{" Apply the given extrinsic.",
				"",
				" Returns an inclusion outcome which specifies if this extrinsic is included in",
				" this block or not."},
		},
		primitives.RuntimeApiMethodMetadata{
			Name:   "finalize_block",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.Header),
			Docs:   sc.Sequence[sc.Str]{" Finish the current block."},
		},
		primitives.RuntimeApiMethodMetadata{
			Name: "inherent_extrinsics",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "inherent",
					Type: sc.ToCompact(metadata.TypesInherentData),
				},
			},
			Output: sc.ToCompact(metadata.TypesSequenceUncheckedExtrinsics),
			Docs:   sc.Sequence[sc.Str]{" Generate inherent extrinsics. The inherent data will vary from chain to chain."},
		},
		primitives.RuntimeApiMethodMetadata{
			Name: "check_inherents",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "block",
					Type: sc.ToCompact(blockId),
				},
				primitives.RuntimeApiMethodParamMetadata{
					Name: "data",
					Type: sc.ToCompact(metadata.TypesInherentData),
				},
			},
			Output: sc.ToCompact(metadata.CheckInherentsResult),
			Docs:   sc.Sequence[sc.Str]{" Check that the inherents are valid. The inherent data will vary from chain to chain."},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{" The `BlockBuilder` api trait that provides the required functionality for building a block."},
	}
}
