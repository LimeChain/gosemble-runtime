package transaction_payment

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/transaction_payment"
	tx_types "github.com/LimeChain/gosemble/frame/transaction_payment/types"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "TransactionPaymentApi"
	apiVersion    = 3
)

// Module implements the TransactionPaymentApi Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-runtime-transactionpaymentapi-module
type Module struct {
	decoder    types.RuntimeDecoder
	txPayments transaction_payment.Module
	memUtils   utils.WasmMemoryTranslator
	logger     log.RuntimeLogger
}

func New(decoder types.RuntimeDecoder, txPayments transaction_payment.Module, logger log.RuntimeLogger) Module {
	return Module{
		decoder:    decoder,
		txPayments: txPayments,
		memUtils:   utils.NewMemoryTranslator(),
		logger:     logger,
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

// QueryInfo queries the data of an extrinsic.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded extrinsic and its length.
// Returns a pointer-size of the SCALE-encoded weight, dispatch class and partial fee.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-transactionpaymentapi-query-info
func (m Module) QueryInfo(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	ext, err := m.decoder.DecodeUncheckedExtrinsic(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	length, err := sc.DecodeU32(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	dispatchInfo := primitives.GetDispatchInfo(ext.Function())

	partialFee := sc.NewU128(0)
	if ext.IsSigned() {
		partialFee, err = m.txPayments.ComputeFee(length, dispatchInfo, constants.DefaultTip)
		if err != nil {
			m.logger.Critical(err.Error())
		}
	}

	runtimeDispatchInfo := primitives.RuntimeDispatchInfo{
		Weight:     dispatchInfo.Weight,
		Class:      dispatchInfo.Class,
		PartialFee: partialFee,
	}

	return m.memUtils.BytesToOffsetAndSize(runtimeDispatchInfo.Bytes())
}

// QueryFeeDetails queries the detailed fee of an extrinsic.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded extrinsic and its length.
// Returns a pointer-size of the SCALE-encoded detailed fee.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-transactionpaymentapi-query-fee-details
func (m Module) QueryFeeDetails(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	ext, err := m.decoder.DecodeUncheckedExtrinsic(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	length, err := sc.DecodeU32(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	dispatchInfo := primitives.GetDispatchInfo(ext.Function())

	var feeDetails tx_types.FeeDetails
	if ext.IsSigned() {
		feeDetails, err = m.txPayments.ComputeFeeDetails(length, dispatchInfo, constants.DefaultTip)
		if err != nil {
			m.logger.Critical(err.Error())
		}
	} else {
		feeDetails = tx_types.FeeDetails{
			InclusionFee: sc.NewOption[tx_types.InclusionFee](nil),
		}
	}

	return m.memUtils.BytesToOffsetAndSize(feeDetails.Bytes())
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name: "query_info",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "uxt",
					Type: sc.ToCompact(metadata.UncheckedExtrinsic),
				},
				primitives.RuntimeApiMethodParamMetadata{
					Name: "len",
					Type: sc.ToCompact(metadata.PrimitiveTypesU32),
				},
			},
			Output: sc.ToCompact(metadata.TypesTransactionPaymentRuntimeDispatchInfo),
			Docs:   sc.Sequence[sc.Str]{},
		},
		primitives.RuntimeApiMethodMetadata{
			Name: "query_fee_details",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "uxt",
					Type: sc.ToCompact(metadata.UncheckedExtrinsic),
				},
				primitives.RuntimeApiMethodParamMetadata{
					Name: "len",
					Type: sc.ToCompact(metadata.PrimitiveTypesU32),
				},
			},
			Output: sc.ToCompact(metadata.TypesTransactionPaymentFeeDetails),
			Docs:   sc.Sequence[sc.Str]{},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{},
	}
}
