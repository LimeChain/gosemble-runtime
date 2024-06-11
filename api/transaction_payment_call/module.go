package transaction_payment_call

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/transaction_payment"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "TransactionPaymentCallApi"
	apiVersion    = 3
)

// Module implements the TransactionPaymentCallApi Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-runtime-transactionpaymentcallapi-module
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

// QueryCallInfo queries the data of a dispatch call.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded dispatch call and its length.
// Returns a pointer-size of the SCALE-encoded weight, dispatch class and partial fee.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-transactionpaymentcallapi-query-call-info
func (m Module) QueryCallInfo(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	call, err := m.decoder.DecodeCall(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	length, err := sc.DecodeU32(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	dispatchInfo := primitives.GetDispatchInfo(call)
	partialFee, err := m.txPayments.ComputeFee(length, dispatchInfo, constants.DefaultTip)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	runtimeDispatchInfo := primitives.RuntimeDispatchInfo{
		Weight:     dispatchInfo.Weight,
		Class:      dispatchInfo.Class,
		PartialFee: partialFee,
	}

	return m.memUtils.BytesToOffsetAndSize(runtimeDispatchInfo.Bytes())
}

// QueryCallFeeDetails queries the detailed fee of a dispatch call.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded dispatch call and its length.
// Returns a pointer-size of the SCALE-encoded detailed fee.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-transactionpaymentcallapi-query-call-fee-details
func (m Module) QueryCallFeeDetails(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	call, err := m.decoder.DecodeCall(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	length, err := sc.DecodeU32(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	dispatchInfo := primitives.GetDispatchInfo(call)
	feeDetails, err := m.txPayments.ComputeFeeDetails(length, dispatchInfo, constants.DefaultTip)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(feeDetails.Bytes())
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name: "query_call_info",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "call",
					Type: sc.ToCompact(metadata.RuntimeCall),
				},
				primitives.RuntimeApiMethodParamMetadata{
					Name: "len",
					Type: sc.ToCompact(metadata.PrimitiveTypesU32),
				},
			},
			Output: sc.ToCompact(metadata.TypesTransactionPaymentRuntimeDispatchInfo),
			Docs:   sc.Sequence[sc.Str]{" Query information of a dispatch class, weight, and fee of a given encoded `Call`."},
		},
		primitives.RuntimeApiMethodMetadata{
			Name: "query_call_fee_details",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{
				primitives.RuntimeApiMethodParamMetadata{
					Name: "call",
					Type: sc.ToCompact(metadata.RuntimeCall),
				},
				primitives.RuntimeApiMethodParamMetadata{
					Name: "len",
					Type: sc.ToCompact(metadata.PrimitiveTypesU32),
				},
			},
			Output: sc.ToCompact(metadata.TypesTransactionPaymentFeeDetails),
			Docs:   sc.Sequence[sc.Str]{" Query fee details of a given encoded `Call`."},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{},
	}
}
