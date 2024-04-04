package session

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type OneSessionHandler interface {
	KeyType() primitives.PublicKeyType
	KeyTypeId() [4]byte
	DecodeKey(buffer *bytes.Buffer) (primitives.Sr25519PublicKey, error)
	OnGenesisSession(validators sc.Sequence[primitives.Validator]) error
	// OnNewSession triggers once the session has changed.
	OnNewSession(changed bool, validators sc.Sequence[primitives.Validator], queuedValidators sc.Sequence[primitives.Validator]) error
	// OnBeforeSessionEnding notifies for the end of the session. Triggered before the end of the session.
	OnBeforeSessionEnding()
	// OnDisabled triggers hook for a disabled validator. Acts accordingly until a new session begins.
	OnDisabled(validatorIndex sc.U32)
}
