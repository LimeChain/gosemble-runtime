package timestamp

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/hooks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	functionSetIndex = iota
	name             = sc.Str("Timestamp")
)

var (
	inherentIdentifier = [8]byte{'t', 'i', 'm', 's', 't', 'a', 'p', '0'}
)

var (
	errTimestampNotUpdated                      = errors.New("Timestamp must be updated once in the block")
	errTimestampInherentNotProvided             = errors.New("Timestamp inherent data must be provided.")
	errTimestampInherentDataNotCorrectlyEncoded = errors.New("Timestamp inherent data not correctly encoded.")
)

type Module struct {
	hooks.DefaultDispatchModule
	Index       sc.U8
	Config      *Config
	storage     *storage
	constants   *consts
	functions   map[sc.U8]primitives.Call
	mdGenerator *primitives.MetadataTypeGenerator
}

func New(index sc.U8, config *Config, mdGenerator *primitives.MetadataTypeGenerator) Module {
	functions := make(map[sc.U8]primitives.Call)
	storage := newStorage(config.Storage)
	constants := newConstants(config.DbWeight, config.MinimumPeriod)
	functions[functionSetIndex] = newCallSet(index, functionSetIndex, storage, constants, config.OnTimestampSet)

	return Module{
		Index:       index,
		Config:      config,
		storage:     storage,
		constants:   constants,
		functions:   functions,
		mdGenerator: mdGenerator,
	}
}

func (m Module) GetIndex() sc.U8 {
	return m.Index
}

func (m Module) name() sc.Str {
	return name
}

func (m Module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m Module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m Module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.DefaultValidTransaction(), nil
}

func (m Module) OnFinalize(_ sc.U64) error {
	value, err := m.storage.DidUpdate.TakeBytes()
	if err != nil {
		return err
	}
	if value == nil {
		return errTimestampNotUpdated
	}
	return nil
}

func (m Module) CreateInherent(inherent primitives.InherentData) (sc.Option[primitives.Call], error) {
	inherentData := inherent.Get(inherentIdentifier)

	if inherentData == nil {
		return sc.Option[primitives.Call]{}, errTimestampInherentNotProvided
	}

	buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(inherentData))
	ts, err := sc.DecodeU64(buffer)
	if err != nil {
		return sc.Option[primitives.Call]{}, errTimestampInherentDataNotCorrectlyEncoded
	}

	now, err := m.storage.Now.Get()
	if err != nil {
		return sc.Option[primitives.Call]{}, err
	}

	nextTimestamp := sc.Max64(ts, now+m.constants.MinimumPeriod)

	function := newCallSetWithArgs(m.Index, functionSetIndex, sc.NewVaryingData(sc.ToCompact(uint64(nextTimestamp))))

	return sc.NewOption[primitives.Call](function), nil
}

func (m Module) CheckInherent(call primitives.Call, inherent primitives.InherentData) error {
	if !m.IsInherent(call) {
		return primitives.NewTimestampErrorInvalid()
	}

	maxTimestampDriftMillis := sc.U64(30 * 1000)

	compactTs := call.Args()[0].(sc.Compact)
	t := sc.U64(compactTs.ToBigInt().Uint64())

	inherentData := inherent.Get(inherentIdentifier)

	if inherentData == nil {
		return errTimestampInherentNotProvided
	}

	buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(inherentData))
	ts, err := sc.DecodeU64(buffer)
	if err != nil {
		return errTimestampInherentDataNotCorrectlyEncoded
	}

	systemNow, err := m.storage.Now.Get()
	if err != nil {
		return err
	}

	minimum := systemNow + m.constants.MinimumPeriod
	if t > ts+maxTimestampDriftMillis {
		return primitives.NewTimestampErrorTooFarInFuture()
	} else if t < minimum {
		return primitives.NewTimestampErrorTooEarly()
	}

	return nil
}

func (m Module) InherentIdentifier() [8]byte {
	return inherentIdentifier
}

func (m Module) IsInherent(call primitives.Call) bool {
	return call.ModuleIndex() == m.Index && call.FunctionIndex() == functionSetIndex
}

func (m Module) Metadata() primitives.MetadataModule {
	timestampCallsMetadataId := m.mdGenerator.BuildCallsMetadata("Timestamp", m.functions, &sc.Sequence[primitives.MetadataTypeParameter]{primitives.NewMetadataEmptyTypeParameter("T")})
	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(timestampCallsMetadataId)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(timestampCallsMetadataId, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Timestamp, Runtime>"),
				},
				m.Index,
				"Call.Timestamp"),
		),
		Event:    sc.NewOption[sc.Compact](nil),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{
			primitives.NewMetadataModuleConstant(
				"MinimumPeriod",
				sc.ToCompact(metadata.PrimitiveTypesU64),
				sc.BytesToSequenceU8(m.constants.MinimumPeriod.Bytes()),
				"The minimum period between blocks. Beware that this is different to the *expected*  period that the block production apparatus provides.",
			),
		},
		Error:    sc.NewOption[sc.Compact](nil),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:    m.Index,
	}

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

func (m Module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
		Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
			primitives.NewMetadataModuleStorageEntry(
				"Now",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU64)),
				"Current time for the current block."),
			primitives.NewMetadataModuleStorageEntry(
				"DidUpdate",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesBool)),
				"Did the timestamp get updated in this block?"),
		},
	})
}
