package session

import (
	"bytes"
	"reflect"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Handler handles session lifecycle events.
type Handler interface {
	// KeyTypeIds returns all the key type ids this session can process.
	KeyTypeIds() sc.Sequence[sc.FixedSequence[sc.U8]]
	// DecodeKeys returns the session keys
	DecodeKeys(buffer *bytes.Buffer) (sc.FixedSequence[primitives.Sr25519PublicKey], error)
	// OnGenesisSession triggers once the genesis session begins. The validator set will be used for the genesis session.
	OnGenesisSession(validators sc.Sequence[queuedKey]) error
	// OnNewSession triggers once the session has changed.
	OnNewSession(changed bool, validators sc.Sequence[queuedKey], queuedValidators sc.Sequence[queuedKey]) error
	// OnBeforeSessionEnding notifies for the end of the session. Triggered before the end of the session.
	OnBeforeSessionEnding()
	// OnDisabled triggers hook for a disabled validator. Acts accordingly until a new session begins.
	OnDisabled(validatorIndex sc.U32)

	AppendHandlers(module OneSessionHandler)
}

type handler struct {
	modules []OneSessionHandler
}

func NewHandler(modules []OneSessionHandler) Handler {
	return handler{modules: modules}
}

// KeyTypeIds returns all the key type ids this session can process.
func (h handler) KeyTypeIds() sc.Sequence[sc.FixedSequence[sc.U8]] {
	var result sc.Sequence[sc.FixedSequence[sc.U8]]

	for _, module := range h.modules {
		keyTypeId := module.KeyTypeId()
		result = append(result, sc.BytesToFixedSequenceU8(keyTypeId[:]))
	}

	return result
}

func (h handler) DecodeKeys(buffer *bytes.Buffer) (sc.FixedSequence[primitives.Sr25519PublicKey], error) {
	var result sc.FixedSequence[primitives.Sr25519PublicKey]

	for _, module := range h.modules {
		key, err := module.DecodeKey(buffer)
		if err != nil {
			return nil, err
		}
		result = append(result, key)
	}

	return result, nil
}

// OnGenesisSession triggers once the genesis session begins. The validator set will be used for the genesis session.
func (h handler) OnGenesisSession(validators sc.Sequence[queuedKey]) error {
	for _, module := range h.modules {
		keys, err := takeOutKeys(module, validators)
		if err != nil {
			return err
		}

		err = module.OnGenesisSession(keys)
		if err != nil {
			return err
		}
	}

	return nil
}

// OnNewSession triggers once the session has changed.
func (h handler) OnNewSession(changed bool, validators sc.Sequence[queuedKey], queuedValidators sc.Sequence[queuedKey]) error {
	for _, module := range h.modules {
		currentKeys, err := takeOutKeys(module, validators)
		if err != nil {
			return err
		}

		nextKeys, err := takeOutKeys(module, queuedValidators)
		if err != nil {
			return err
		}

		err = module.OnNewSession(changed, currentKeys, nextKeys)
		if err != nil {
			return err
		}
	}
	return nil
}

// OnBeforeSessionEnding notifies for the end of the session. Triggered before the end of the session.
func (h handler) OnBeforeSessionEnding() {
	for _, module := range h.modules {
		module.OnBeforeSessionEnding()
	}
}

// OnDisabled triggers hook for a disabled validator. Acts accordingly until a new session begins.
func (h handler) OnDisabled(validatorIndex sc.U32) {
	for _, module := range h.modules {
		module.OnDisabled(validatorIndex)
	}
}

func takeOutKeys(module OneSessionHandler, validators sc.Sequence[queuedKey]) (sc.Sequence[primitives.Validator], error) {
	keyTypeId := module.KeyTypeId()
	fixedKeyTypeId := sc.BytesToFixedSequenceU8(keyTypeId[:])

	var result sc.Sequence[primitives.Validator]
	for _, validator := range validators {
		for _, sessionKey := range validator.Keys {
			key, err := primitives.NewSr25519PublicKey(sessionKey.Key...)
			if err != nil {
				return nil, err
			}

			if reflect.DeepEqual(sessionKey.TypeId, fixedKeyTypeId) {
				result = append(result, primitives.Validator{
					AccountId:   validator.Validator,
					AuthorityId: key,
				})
			}
		}
	}

	return result, nil
}

func (h handler) AppendHandlers(module OneSessionHandler) {
	h.modules = append(h.modules, module)
}
