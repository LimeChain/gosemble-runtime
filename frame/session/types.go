package session

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type keyManager interface {
	DoSetKeys(who primitives.AccountId, sessionKeys sc.Sequence[primitives.SessionKey]) error
	DoPurgeKeys(who primitives.AccountId) error
}

type queuedKey struct {
	Validator primitives.AccountId
	Keys      sc.Sequence[primitives.SessionKey]
}

func (qk queuedKey) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, qk.Validator, qk.Keys)
}

func DecodeQueuedKey(buffer *bytes.Buffer) (queuedKey, error) {
	accId, err := primitives.DecodeAccountId(buffer)
	if err != nil {
		return queuedKey{}, err
	}
	keys, err := sc.DecodeSequenceWith(buffer, primitives.DecodeSessionKey)
	if err != nil {
		return queuedKey{}, err
	}

	return queuedKey{
		Validator: accId,
		Keys:      keys,
	}, nil
}

func (qk queuedKey) Bytes() []byte {
	return sc.EncodedBytes(qk)
}

func DecodeQueuedKeys(buffer *bytes.Buffer) (sc.Sequence[queuedKey], error) {
	return sc.DecodeSequenceWith(buffer, DecodeQueuedKey)
}
