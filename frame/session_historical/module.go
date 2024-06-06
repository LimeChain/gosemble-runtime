package session_historical

import (
	"bytes"

	"github.com/LimeChain/gosemble/frame/session"
	"github.com/LimeChain/gosemble/primitives/log"
	sessiontypes "github.com/LimeChain/gosemble/primitives/session"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/hooks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const name = sc.Str("SessionHistorical")

type Module interface {
	primitives.Module

	Prove(key [4]byte, authorityId primitives.AccountId) sc.Option[sessiontypes.MembershipProof]
	CheckProof(key [4]byte, authorityId primitives.AccountId, proof sessiontypes.MembershipProof) sc.Option[IdentificationTuple]
}

type module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index         sc.U8
	config        *Config
	storage       *storage
	constants     *consts
	sessionModule session.Module
	mdGenerator   *primitives.MetadataTypeGenerator
	logger        log.Logger
}

func New(index sc.U8, config *Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.Logger) Module {
	storage := newStorage()

	return module{
		index:         index,
		config:        config,
		storage:       storage,
		constants:     newConstants(),
		sessionModule: config.SessionModule,
		mdGenerator:   mdGenerator,
		logger:        logger,
	}
}

func (m module) GetIndex() sc.U8 {
	return m.index
}

func (m module) name() sc.Str {
	return name
}

func (m module) Functions() map[sc.U8]primitives.Call {
	return map[sc.U8]primitives.Call{}
}

func (m module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

func (m module) DecodeKey(buffer *bytes.Buffer) (primitives.Sr25519PublicKey, error) {
	key, err := primitives.DecodeSr25519PublicKey(buffer)
	if err != nil {
		return primitives.Sr25519PublicKey{}, err
	}

	return key, nil
}

func (m module) OnGenesisSession(validators sc.Sequence[primitives.Validator]) error {
	return nil
}

func (m module) OnNewSession(isChanged bool, validators sc.Sequence[primitives.Validator], _ sc.Sequence[primitives.Validator]) error {
	return nil
}

func (m module) OnBeforeSessionEnding() {}

func (m module) OnDisabled(validatorIndex sc.U32) {}

func (m module) Metadata() primitives.MetadataModule {
	dataV14 := primitives.MetadataModuleV14{
		Name:      m.name(),
		Storage:   sc.Option[primitives.MetadataModuleStorage]{},
		Call:      sc.NewOption[sc.Compact](nil),
		CallDef:   sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:     m.index,
	}

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

// A trie instance for checking and generating proofs.
type ProvingTrie struct {
	// db   MemoryDB // TODO
	root primitives.H256
}

// KeyOwnerProofSystem interface

func (m module) Prove(key [4]byte, authorityId primitives.AccountId) sc.Option[sessiontypes.MembershipProof] {
	// sessionIndex, err := m.sessionModule.CurrentIndex()
	// if err != nil {
	// 	return sc.NewOption[sessiontypes.MembershipProof](nil)
	// }

	// validators, err := m.sessionModule.Validators()
	// if err != nil {
	// 	return sc.NewOption[sessiontypes.MembershipProof](nil)
	// }

	// TODO: once we implement staking
	// 	validators.into_iter()
	// 	.filter_map(|validator| {
	// 		T::FullIdentificationOf::convert(validator.clone())
	// 			.map(|full_id| (validator, full_id))
	// 	})
	// 	.collect::<Vec<_>>();

	// count := sc.U32(len(validators))

	// 	trie, err := ProvingTrieGenerateFor(validators)
	// 	if err != nil {
	// 		return sc.NewOption[session.MembershipProof](nil)
	// 	}

	// 	// (id, data) = key

	//	trie.prove(id, data.as_ref()).map(|trie_nodes| MembershipProof {
	//		sessionIndex,
	//		trie_nodes,
	//		validator_count: count,
	//	})

	return sc.NewOption[sessiontypes.MembershipProof](nil)
}

type IdentificationTuple = sc.Bool

func (m module) CheckProof(key [4]byte, authorityId primitives.AccountId, proof sessiontypes.MembershipProof) sc.Option[IdentificationTuple] {
	// TODO:
	// if proof.session == <Session<T>>::current_index() {
	// 	<Session<T>>::key_owner(key, authorityId.as_ref()).and_then(|owner| {
	// 		T::FullIdentificationOf::convert(owner.clone()).and_then(move |id| {
	// 			let count = <Session<T>>::validators().len() as ValidatorCount;

	// 			if count != proof.validator_count {
	// 				return None
	// 			}

	// 			Some((owner, key))
	// 		})
	// 	})
	// } else {
	// 	let (root, count) = <HistoricalSessions<T>>::get(&proof.session)?;

	// 	if count != proof.validator_count {
	// 		return None
	// 	}

	// 	let trie = ProvingTrie::<T>::from_nodes(root, &proof.trie_nodes);
	// 	trie.query(key, authorityId.as_ref())
	// }

	return sc.NewOption[IdentificationTuple](nil)
}
