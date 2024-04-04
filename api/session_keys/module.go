package session_keys

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "SessionKeys"
	apiVersion    = 1
)

// Module implements the SessionKeys Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-runtime-sessionkeys-module
type Module struct {
	sessions []types.Session
	crypto   io.Crypto
	memUtils utils.WasmMemoryTranslator
	logger   log.Logger
}

func New(sessions []types.Session, logger log.Logger) Module {
	return Module{
		sessions: sessions,
		crypto:   io.NewCrypto(),
		memUtils: utils.NewMemoryTranslator(),
		logger:   logger,
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

// GenerateSessionKeys generates a set of session keys with an optional seed.
// The keys should be stored within the keystore exposed by the Host Api.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded optional seed.
// Returns a pointer-size of the SCALE-encoded set of keys.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#id-sessionkeys_generate_session_keys
func (m Module) GenerateSessionKeys(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	seed, err := sc.DecodeOptionWith(buffer, sc.DecodeSequence[sc.U8])
	if err != nil {
		m.logger.Critical(err.Error())
	}

	var publicKeys []byte
	for _, session := range m.sessions {
		keyGenerationFunc := getKeyFunction(m, session.KeyType())
		keyTypeId := session.KeyTypeId()

		publicKey := keyGenerationFunc(keyTypeId[:], seed.Bytes())
		publicKeys = append(publicKeys, publicKey...)
	}

	res := sc.BytesToSequenceU8(publicKeys)

	return m.memUtils.BytesToOffsetAndSize(res.Bytes())
}

// DecodeSessionKeys decodes the given session keys.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded keys.
// Returns a pointer-size of the SCALE-encoded set of raw keys and their respective key type.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#id-sessionkeys_decode_session_keys
func (m Module) DecodeSessionKeys(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)
	sequence, err := sc.DecodeSequenceWith(buffer, sc.DecodeU8)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	buffer = bytes.NewBuffer(sc.SequenceU8ToBytes(sequence))
	sessionKeys := sc.Sequence[types.SessionKey]{}
	for _, session := range m.sessions {
		pk, err := types.DecodeAccountId(buffer)
		if err != nil {
			m.logger.Critical(err.Error())
		}
		sessionKey := types.NewSessionKey(pk.Bytes(), session.KeyTypeId())
		sessionKeys = append(sessionKeys, sessionKey)
	}

	result := sc.NewOption[sc.Sequence[types.SessionKey]](sessionKeys)
	return m.memUtils.BytesToOffsetAndSize(result.Bytes())
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() types.RuntimeApiMetadata {
	methods := sc.Sequence[types.RuntimeApiMethodMetadata]{
		types.RuntimeApiMethodMetadata{
			Name: "generate_session_keys",
			Inputs: sc.Sequence[types.RuntimeApiMethodParamMetadata]{
				types.RuntimeApiMethodParamMetadata{
					Name: "seed",
					Type: sc.ToCompact(metadata.TypesOptionSequenceU8),
				},
			},
			Output: sc.ToCompact(metadata.TypesSequenceU8),
			Docs: sc.Sequence[sc.Str]{
				" Generate a set of session keys with optionally using the given seed.",
				" The keys should be stored within the keystore exposed via runtime",
				" externalities.",
				"",
				" The seed needs to be a valid `utf8` string.",
				"",
				" Returns the concatenated SCALE encoded public keys.",
			},
		},
		types.RuntimeApiMethodMetadata{
			Name: "decode_session_keys",
			Inputs: sc.Sequence[types.RuntimeApiMethodParamMetadata]{
				types.RuntimeApiMethodParamMetadata{
					Name: "encoded",
					Type: sc.ToCompact(metadata.TypesSequenceU8),
				},
			},
			Output: sc.ToCompact(metadata.TypesOptionTupleSequenceU8KeyTypeId),
			Docs: sc.Sequence[sc.Str]{
				" Decode the given public session keys.",
				"",
				" Returns the list of public raw public keys + key type.",
			},
		},
	}

	return types.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs:    sc.Sequence[sc.Str]{" Session keys runtime api."},
	}
}

func getKeyFunction(m Module, keyType types.PublicKeyType) func([]byte, []byte) []byte {
	switch keyType {
	case types.PublicKeyEd25519:
		return m.crypto.Ed25519Generate
	case types.PublicKeySr25519:
		return m.crypto.Sr25519Generate
	case types.PublicKeyEcdsa:
		return m.crypto.EcdsaGenerate
	default:
		m.logger.Critical("invalid public key type")
	}

	panic("unreachable")
}
