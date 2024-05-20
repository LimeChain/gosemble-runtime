package authorship

import (
	"bytes"

	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/log"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/hooks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const name = sc.Str("Authorship")

type Module interface {
	primitives.Module

	Author() sc.Option[primitives.AccountId]
}

type module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index        sc.U8
	config       *Config
	storage      *storage
	constants    *consts
	functions    map[sc.U8]primitives.Call
	systemModule system.Module
	mdGenerator  *primitives.MetadataTypeGenerator
	logger       log.Logger
}

func New(index sc.U8, config *Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.Logger) Module {
	storage := newStorage()

	return module{
		index:       index,
		config:      config,
		storage:     storage,
		constants:   newConstants(),
		mdGenerator: mdGenerator,
		logger:      logger,
	}
}

func (m module) GetIndex() sc.U8 {
	return m.index
}

func (m module) name() sc.Str {
	return name
}

func (m module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

func (m module) OnInitialize(_ sc.U64) (primitives.Weight, error) {
	if author := m.Author(); author.HasValue {
		m.config.EventHandler.NoteAuthor(author.Value)
	}
	return primitives.WeightZero(), nil
}

func (m module) OnFinalize(_ sc.U64) error {
	// ensure we never go to trie with these values.
	m.storage.Author.Clear()
	return nil
}

// Fetch the author of the block.
//
// This is safe to invoke in `on_initialize` implementations, as well
// as afterwards.
func (m module) Author() sc.Option[primitives.AccountId] {
	// Check the memorized storage value.
	author, err := m.storage.Author.GetBytes()
	if err != nil {
		return sc.NewOption[primitives.AccountId](nil)
	}

	if author.HasValue {
		author, err := primitives.DecodeAccountId(bytes.NewBuffer(sc.SequenceU8ToBytes(author.Value)))
		if err != nil {
			return sc.NewOption[primitives.AccountId](nil)
		}
		return sc.NewOption[primitives.AccountId](author)
	}

	digest, err := m.config.SystemModule.StorageDigest()
	if err != nil {
		return sc.NewOption[primitives.AccountId](nil)
	}

	preRuntimeDigests, err := digest.PreRuntimes()
	if err != nil {
		return sc.NewOption[primitives.AccountId](nil)
	}

	authorId, err := m.config.FindAuthor.FindAuthor(preRuntimeDigests)
	if err != nil {
		return sc.NewOption[primitives.AccountId](nil)
	}

	m.storage.Author.Put(authorId.Value)
	return authorId
}

type EventHandler interface {
	NoteAuthor(author primitives.AccountId)
}

type DefaulthEventHandler struct{}

func (d DefaulthEventHandler) NoteAuthor(author primitives.AccountId) {}
