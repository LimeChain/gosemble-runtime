package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type SessionKey struct {
	Key    sc.Sequence[sc.U8]
	TypeId sc.FixedSequence[sc.U8]
}

func NewSessionKey(key []byte, typeId [4]byte) SessionKey {
	return SessionKey{
		Key:    sc.BytesToSequenceU8(key),
		TypeId: sc.BytesToFixedSequenceU8(typeId[:]),
	}
}

func NewSessionKeyFromBytes(key []byte, typeId sc.FixedSequence[sc.U8]) SessionKey {
	return SessionKey{
		Key:    sc.BytesToSequenceU8(key),
		TypeId: typeId,
	}
}

func (sk SessionKey) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, sk.Key, sk.TypeId)
}

func DecodeSessionKey(buffer *bytes.Buffer) (SessionKey, error) {
	key, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return SessionKey{}, err
	}
	typeId, err := sc.DecodeFixedSequence[sc.U8](4, buffer)
	if err != nil {
		return SessionKey{}, err
	}
	return SessionKey{
		Key:    key,
		TypeId: typeId,
	}, nil
}

func (sk SessionKey) Bytes() []byte {
	return sc.EncodedBytes(sk)
}
