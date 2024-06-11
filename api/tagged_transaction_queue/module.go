package tagged_transaction_queue

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
	ApiModuleName = "TaggedTransactionQueue"
	apiVersion    = 3
)

type TaggedTransactionQueue interface {
	ValidateTransaction(dataPtr int32, dataLen int32) int64
}

// Module implements the TaggedTransactionQueue Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-runtime-txqueue-module
type Module struct {
	executive   executive.Module
	decoder     types.RuntimeDecoder
	memUtils    utils.WasmMemoryTranslator
	mdGenerator *primitives.MetadataTypeGenerator
	logger      log.RuntimeLogger
}

func New(executive executive.Module, decoder types.RuntimeDecoder, mdGenerator *primitives.MetadataTypeGenerator, logger log.RuntimeLogger) Module {
	return Module{
		executive:   executive,
		decoder:     decoder,
		memUtils:    utils.NewMemoryTranslator(),
		mdGenerator: mdGenerator,
		logger:      logger,
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

// ValidateTransaction validates an extrinsic at a given block.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded tx source, extrinsic and block hash.
// Returns a pointer-size of the SCALE-encoded result whether the extrinsic is valid.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-validate-transaction
func (m Module) ValidateTransaction(dataPtr int32, dataLen int32) int64 {
	data := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(data)

	txSource, err := primitives.DecodeTransactionSource(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	tx, err := m.decoder.DecodeUncheckedExtrinsic(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	blockHash, err := primitives.DecodeBlake2bHash(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	ok, err := m.executive.ValidateTransaction(txSource, tx, blockHash)

	var res primitives.TransactionValidityResult
	switch typedErr := err.(type) {
	case primitives.TransactionValidityError:
		res, err = primitives.NewTransactionValidityResult(typedErr)
	case nil:
		res, err = primitives.NewTransactionValidityResult(ok)
	}
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(res.Bytes())
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	transactionSourceId, _ := m.mdGenerator.GetId("TransactionSource")
	resultValidityTxId, _ := m.mdGenerator.GetId("TransactionValidityResult")

	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name: "validate_transaction",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "source",
					Type: sc.ToCompact(transactionSourceId),
				},
				primitives.RuntimeApiMethodParamMetadata{
					Name: "tx",
					Type: sc.ToCompact(metadata.UncheckedExtrinsic),
				},
				primitives.RuntimeApiMethodParamMetadata{
					Name: "block_hash",
					Type: sc.ToCompact(metadata.TypesH256),
				},
			},
			Output: sc.ToCompact(resultValidityTxId),
			Docs: sc.Sequence[sc.Str]{" Validate the transaction.",
				"",
				" This method is invoked by the transaction pool to learn details about given transaction.",
				" The implementation should make sure to verify the correctness of the transaction",
				" against current state. The given `block_hash` corresponds to the hash of the block",
				" that is used as current state.",
				"",
				" Note that this call may be performed by the pool multiple times and transactions",
				" might be verified in any possible order."},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{" The `TaggedTransactionQueue` api trait for interfering with the transaction queue."},
	}
}
