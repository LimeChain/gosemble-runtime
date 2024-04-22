package grandpa

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	name = sc.Str("Grandpa")
)

const (
	PauseFailedError = iota
	ResumeFailedError
	ChangePendingError
	TooSoonError
	InvalidKeyOwnershipProofError
	InvalidEquivocationProofError
	DuplicateOffenceReportError
)

var (
	AuthorityVersion sc.U8 = 1
	EngineId               = [4]byte{'f', 'r', 'n', 'k'}
	KeyTypeId              = [4]byte{'g', 'r', 'a', 'n'}
)

type GrandpaModule interface {
	primitives.Module

	KeyType() primitives.PublicKeyType
	KeyTypeId() [4]byte
	Authorities() (sc.Sequence[primitives.Authority], error)
}

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	Index       sc.U8
	storage     *storage
	functions   map[sc.U8]primitives.Call
	mdGenerator *primitives.MetadataTypeGenerator
	logger      log.WarnLogger
}

func New(index sc.U8, logger log.WarnLogger, mdGenerator *primitives.MetadataTypeGenerator) Module {

	return Module{
		Index:       index,
		storage:     newStorage(),
		mdGenerator: mdGenerator,
		logger:      logger,
	}
}

func (m Module) KeyType() primitives.PublicKeyType {
	return primitives.PublicKeyEd25519
}

func (m Module) KeyTypeId() [4]byte {
	return KeyTypeId
}

func (m Module) GetIndex() sc.U8 {
	return m.Index
}

func (m Module) name() sc.Str {
	return name
}

func (m Module) Functions() map[sc.U8]primitives.Call {
	return map[sc.U8]primitives.Call{}
}

func (m Module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m Module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

func (m Module) Authorities() (sc.Sequence[primitives.Authority], error) {
	return m.storage.Authorities.Get()
}

func (m Module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
		Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
			primitives.NewMetadataModuleStorageEntry(
				"Authorities",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesBoundedVecAuthority)),
				"The current list of authorities.",
			),
		},
	})
}

func (m Module) Metadata() primitives.MetadataModule {
	m.mdGenerator.BuildCallsMetadata("Grandpa", m.Functions(), &sc.Sequence[primitives.MetadataTypeParameter]{
		primitives.NewMetadataEmptyTypeParameter("T"),
		primitives.NewMetadataEmptyTypeParameter("I"),
	})

	dataV14 := primitives.MetadataModuleV14{
		Name:      m.name(),
		Storage:   sc.Option[primitives.MetadataModuleStorage]{},
		Call:      sc.NewOption[sc.Compact](nil),
		CallDef:   sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesGrandpaErrors),
				},
				m.Index,
				"Errors.Grandpa"),
		),
		Index: m.Index,
	}
	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

func (m Module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithParams(metadata.TypesGrandpaErrors, "The `Error` enum of this pallet.", sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Error"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant("PauseFailed", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, PauseFailedError, ""),
				primitives.NewMetadataDefinitionVariant("ResumeFailed", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ResumeFailedError, ""),
				primitives.NewMetadataDefinitionVariant("ChangePending", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ChangePendingError, ""),
				primitives.NewMetadataDefinitionVariant("TooSoon", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, TooSoonError, ""),
				primitives.NewMetadataDefinitionVariant("InvalidKeyOwnershipProof", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, InvalidKeyOwnershipProofError, ""),
				primitives.NewMetadataDefinitionVariant("InvalidEquivocationProof", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, InvalidEquivocationProofError, ""),
				primitives.NewMetadataDefinitionVariant("DuplicateOffenceReport", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, DuplicateOffenceReportError, ""),
			}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
			}),
		primitives.NewMetadataTypeWithPath(metadata.TypesGrandpaAppPublic, "sp_consensus_grandpa app Public", sc.Sequence[sc.Str]{"sp_consensus_grandpa", "app", "Public"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionField(metadata.TypesEd25519PubKey),
			})),
		primitives.NewMetadataType(metadata.TypesTupleGrandpaAppPublicU64, "(GrandpaAppPublic, U64)",
			primitives.NewMetadataTypeDefinitionTuple(sc.Sequence[sc.Compact]{sc.ToCompact(metadata.TypesGrandpaAppPublic), sc.ToCompact(metadata.PrimitiveTypesU64)})),
		primitives.NewMetadataType(metadata.TypesSequenceTupleGrandpaAppPublic, "[]byte (GrandpaAppPublic, U64)", primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleGrandpaAppPublicU64))),
	}
}
