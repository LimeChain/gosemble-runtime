package sudo

import (
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	functionSudo = iota
	functionSudoUncheckedWeight
	functionSetKey
	functionSudoAs
	functionRemoveKey
)

const (
	name = sc.Str("Sudo")
)

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index          sc.U8
	functions      map[sc.U8]primitives.Call
	mdGenerator    *primitives.MetadataTypeGenerator
	storage        *storage
	eventDepositor primitives.EventDepositor
	logger         log.RuntimeLogger
}

func New(index sc.U8, config Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.RuntimeLogger) Module {
	functions := make(map[sc.U8]primitives.Call)

	module := Module{
		index:          index,
		storage:        newStorage(),
		eventDepositor: config.EventDepositor,
		mdGenerator:    mdGenerator,
		logger:         logger,
	}

	functions[functionSudo] = newCallSudo(index, functionSudo, config.DbWeight, module)
	functions[functionSudoUncheckedWeight] = newCallSudoUncheckedWeight(index, functionSudoUncheckedWeight, config.DbWeight, module)
	functions[functionSetKey] = newCallSetKey(index, functionSetKey, config.DbWeight, module)
	functions[functionSudoAs] = newCallSudoAs(index, functionSudoAs, config.DbWeight, module)
	functions[functionRemoveKey] = newCallRemoveKey(index, functionRemoveKey, config.DbWeight, module)

	module.functions = functions

	return module
}

func (m Module) name() sc.Str {
	return name
}

func (m Module) GetIndex() sc.U8 { return m.index }

func (m Module) Functions() map[sc.U8]primitives.Call { return m.functions }

func (m Module) PreDispatch(call primitives.Call) (sc.Empty, error) { return sc.Empty{}, nil }

func (m Module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

func (m *Module) executeCall(origin primitives.RuntimeOrigin, call primitives.Call, eventFunc func(u8 sc.U8, outcome primitives.DispatchOutcome) primitives.Event) (primitives.PostDispatchInfo, error) {
	_, err := call.Dispatch(origin, call.Args())
	var outcome primitives.DispatchOutcome
	if err != nil {
		dispatchErr, ok := err.(primitives.DispatchError)
		if !ok {
			return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		}

		outcome, err = primitives.NewDispatchOutcome(dispatchErr)
	} else {
		outcome, err = primitives.NewDispatchOutcome(sc.Empty{})
	}

	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	m.eventDepositor.DepositEvent(eventFunc(m.index, outcome))

	return primitives.PostDispatchInfo{
		PaysFee: primitives.PaysNo,
	}, nil
}

func (m Module) ensureSudo(origin primitives.RuntimeOrigin) error {
	res, err := system.EnsureSignedOrRoot(origin)
	if err != nil {
		return err
	}

	if res.HasValue {
		key, err := m.storage.Key.Get()
		if err != nil {
			return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
		}
		if !reflect.DeepEqual(res.Value, key) {
			return NewDispatchErrorRequireSudo(m.index)
		}
	}

	return nil
}

func (m Module) Metadata() primitives.MetadataModule {
	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSudoCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSudoCalls, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Sudo, Runtime>"),
				},
				m.index,
				"Call.Sudo")),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSudoEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSudoEvent, "frame_sudo::Event<Runtime>"),
				},
				m.index,
				"Events.System"),
		),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSudoErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesSudoErrors),
				},
				m.index,
				"Errors.Sudo"),
		),
		Index: m.index,
	}

	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

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
				"Key",
				primitives.MetadataModuleStorageEntryModifierOptional,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAddress32)),
				"The `AccountId` of the sudo key.",
			),
		},
	})
}

func (m Module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithPath(
			metadata.TypesSudoEvent,
			"pallet_sudo pallet Sudo",
			sc.Sequence[sc.Str]{"pallet_sudo", "pallet", "Event"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Sudid",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchOutcome, "sudo_result", "DispatchResult"),
						},
						EventSudid,
						"Events.Sudid"),
					primitives.NewMetadataDefinitionVariant(
						"KeyChanged",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionAccountId, "old", "Option<T::AccountId>"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "new", "T::AccountId"),
						},
						EventKeyChanged,
						"Events.KeyChanged"),
					primitives.NewMetadataDefinitionVariant(
						"KeyRemoved",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventKeyRemoved,
						"Events.KeyRemoved"),
					primitives.NewMetadataDefinitionVariant(
						"SudoAsDone",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchOutcome, "sudo_result", "DispatchResult"),
						},
						EventSudoAsDone,
						"Events.SudoAsDone"),
				})),

		primitives.NewMetadataTypeWithParams(metadata.TypesSudoErrors,
			"pallet_sudo pallet Error",
			sc.Sequence[sc.Str]{"pallet_sudo", "pallet", "Error"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"RequireSudo",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorRequireSudo,
						"Sender must be the sudo account.",
					),
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			}),

		primitives.NewMetadataTypeWithParam(metadata.TypesSudoCalls,
			"Sudo calls",
			sc.Sequence[sc.Str]{"pallet_sudo", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"sudo",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.RuntimeCall, "call", "Box<<T as Config>::RuntimeCall>"),
						},
						functionSudo,
						"Authenticates the sudo key and dispatches a function call with `Root` origin."),
					primitives.NewMetadataDefinitionVariant(
						"sudo_unchecked_weight",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.RuntimeCall, "call", "Box<<T as Config>::RuntimeCall>"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesWeight, "weight", "Weight"),
						},
						functionSudoUncheckedWeight,
						"Authenticates the sudo key and dispatches a function call with `Root` origin."+
							"This function does not check the weight of the call, and instead allows the"+
							"Sudo user to specify the weight of the call."+
							"The dispatch origin for this call must be `Signed`.",
					),
					primitives.NewMetadataDefinitionVariant(
						"set_key",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "new", "AccountIdLookupOf<T>"),
						},
						functionSetKey,
						"Authenticates the current sudo key and sets the given `AccountId` as the new sudo."),
					primitives.NewMetadataDefinitionVariant(
						"sudo_as",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "who", "AccountIdLookupOf<T>"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.RuntimeCall, "call", "Box<<T as Config>::RuntimeCall>"),
						},
						functionSudoAs,
						"Authenticates the sudo key and dispatches a function call with `Signed` origin from a given account."+
							"The dispatch origin for this call must be `Signed`.",
					),
					primitives.NewMetadataDefinitionVariant(
						"remove_key",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						functionRemoveKey,
						"Permanently removes the sudo key. This cannot be undone."),
				}),
			primitives.NewMetadataEmptyTypeParameter("T")),
	}
}
