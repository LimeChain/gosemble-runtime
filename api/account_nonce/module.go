package account_nonce

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "AccountNonceApi"
	apiVersion    = 1
)

// Module implements the AccountNonceApi Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#id-module-accountnonceapi
type Module struct {
	systemModule system.Module
	memUtils     utils.WasmMemoryTranslator
	logger       log.RuntimeLogger
}

func New(systemModule system.Module, logger log.RuntimeLogger) Module {
	return Module{
		systemModule: systemModule,
		memUtils:     utils.NewMemoryTranslator(),
		logger:       logger,
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

// AccountNonce returns the account nonce of given AccountId.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded AccountId.
// Returns a pointer-size of the SCALE-encoded nonce of the AccountId.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-accountnonceapi-account-nonce
func (m Module) AccountNonce(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	accountId, err := types.DecodeAccountId(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	account, err := m.systemModule.Get(accountId)
	if err != nil {
		m.logger.Critical(err.Error())
	}
	nonce := account.Nonce

	return m.memUtils.BytesToOffsetAndSize(nonce.Bytes())
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() types.RuntimeApiMetadata {
	methods := sc.Sequence[types.RuntimeApiMethodMetadata]{
		types.RuntimeApiMethodMetadata{
			Name: "account_nonce",
			Inputs: sc.Sequence[types.RuntimeApiMethodParamMetadata]{
				types.RuntimeApiMethodParamMetadata{
					Name: "account",
					Type: sc.ToCompact(metadata.TypesAddress32),
				},
			},
			Output: sc.ToCompact(metadata.PrimitiveTypesU32),
			Docs:   sc.Sequence[sc.Str]{" Get current account nonce of given `AccountId`."},
		},
	}

	return types.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{" The API to query account nonce."},
	}
}
