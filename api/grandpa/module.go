package grandpa

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/grandpa"
	grandpatypes "github.com/LimeChain/gosemble/primitives/grandpa"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/session"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "GrandpaApi"
	apiVersion    = 3
)

// Module implements the GrandpaApi Runtime API definition.
//
// For more information about API definition, see:
// https://spec.polkadot.network/chap-runtime-api#id-module-grandpaapi
type Module struct {
	grandpa  grandpa.Module
	memUtils utils.WasmMemoryTranslator
	logger   log.Logger
}

func New(grandpa grandpa.Module, logger log.Logger) Module {
	return Module{
		grandpa:  grandpa,
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

// Authorities returns the current set of Grandpa authorities.
// Returns a pointer-size of the SCALE-encoded set of authorities.
//
// For more information about function definition, see:
// https://spec.polkadot.network/chap-runtime-api#sect-rte-grandpa-auth
func (m Module) Authorities() int64 {
	authorities, err := m.grandpa.Authorities()
	if err != nil {
		m.logger.Critical(err.Error())
	}
	return m.memUtils.BytesToOffsetAndSize(authorities.Bytes())
}

func (m Module) CurrentSetId() int64 {
	setId, err := m.grandpa.StorageSetId()
	if err != nil {
		m.logger.Critical(err.Error())
	}
	return m.memUtils.BytesToOffsetAndSize(setId.Bytes())
}

func (m Module) SubmitReportEquivocationUnsignedExtrinsic(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	equivocationProof, err := grandpatypes.DecodeEquivocationProof(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	opaqueKeyOwnershipProof, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	keyOwnerProof, err := session.DecodeMembershipProof(bytes.NewBuffer(sc.SequenceU8ToBytes(opaqueKeyOwnershipProof)))
	if err != nil {
		m.logger.Critical(err.Error())
	}

	err = m.grandpa.SubmitUnsignedEquivocationReport(equivocationProof, keyOwnerProof)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	return m.memUtils.BytesToOffsetAndSize(sc.NewOption[sc.Empty](nil).Bytes())
}

func (m Module) GenerateKeyOwnershipProof(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	_, err := sc.DecodeU64(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	authorityId, err := primitives.DecodeAccountId(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	res := m.grandpa.HistoricalKeyOwnershipProof(authorityId)
	return m.memUtils.BytesToOffsetAndSize(res.Bytes())

}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() primitives.RuntimeApiMetadata {
	methods := sc.Sequence[primitives.RuntimeApiMethodMetadata]{
		primitives.RuntimeApiMethodMetadata{
			Name:   "grandpa_authorities",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.TypesSequenceTupleGrandpaAppPublic),
			Docs: sc.Sequence[sc.Str]{
				" Get the current GRANDPA authorities and weights. This should not change except",
				" for when changes are scheduled and the corresponding delay has passed.",
				"",
				" When called at block B, it will return the set of authorities that should be",
				" used to finalize descendants of this block (B+1, B+2, ...). The block B itself",
				" is finalized by the authorities from block B-1.",
			},
		},

		primitives.RuntimeApiMethodMetadata{
			Name:   "current_set_id",
			Inputs: sc.Sequence[primitives.RuntimeApiMethodParamMetadata]{},
			Output: sc.ToCompact(metadata.PrimitiveTypesU64),
			Docs: sc.Sequence[sc.Str]{
				"Get current GRANDPA authority set id.",
			},
		},
	}

	return primitives.RuntimeApiMetadata{
		Name:    ApiModuleName,
		Methods: methods,
		Docs: sc.Sequence[sc.Str]{
			" APIs for integrating the GRANDPA finality gadget into runtimes.",
			" This should be implemented on the runtime side.",
			"",
			" This is primarily used for negotiating authority-set changes for the",
			" gadget. GRANDPA uses a signaling model of changing authority sets:",
			" changes should be signaled with a delay of N blocks, and then automatically",
			" applied in the runtime after those N blocks have passed.",
			"",
			" The consensus protocol will coordinate the handoff externally.",
		},
	}
}
