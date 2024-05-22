package testable

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	functionTestIndex = iota
)

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	Index       sc.U8
	functions   map[sc.U8]primitives.Call
	mdGenerator *primitives.MetadataTypeGenerator
}

func New(index sc.U8, ioStorage io.Storage, ioTransactionBroker io.TransactionBroker, mdGenerator *primitives.MetadataTypeGenerator) Module {
	functions := make(map[sc.U8]primitives.Call)
	functions[functionTestIndex] = newCallTest(index, functionTestIndex, ioStorage, ioTransactionBroker)

	return Module{
		Index:       index,
		functions:   functions,
		mdGenerator: mdGenerator,
	}
}

func (m Module) GetIndex() sc.U8 {
	return m.Index
}

func (m Module) name() sc.Str {
	return "Testable"
}

func (m Module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m Module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m Module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

func (m Module) Metadata() primitives.MetadataModule {
	testableCallsMetadataId := m.mdGenerator.BuildCallsMetadata("Testable", m.functions, &sc.Sequence[primitives.MetadataTypeParameter]{primitives.NewMetadataEmptyTypeParameter("T")})

	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: sc.Option[primitives.MetadataModuleStorage]{},
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(testableCallsMetadataId)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(testableCallsMetadataId, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Testable, Runtime>"),
				},
				m.Index,
				"Call.Testable"),
		),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:     m.Index,
	}

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}
