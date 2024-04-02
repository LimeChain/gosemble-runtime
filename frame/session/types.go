package session

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Manager manages the creation of a new validator set.
type Manager interface {
	NewSession(newIndex sc.U32) sc.Option[sc.Sequence[primitives.AccountId]]
	NewSessionGenesis(newIndex sc.U32) sc.Option[sc.Sequence[primitives.AccountId]]
	EndSession(index sc.U32)
	StartSession(index sc.U32)
}

type DefaultManager struct{}

func (dm DefaultManager) NewSession(newIndex sc.U32) sc.Option[sc.Sequence[primitives.AccountId]] {
	return sc.Option[sc.Sequence[primitives.AccountId]]{}
}
func (dm DefaultManager) NewSessionGenesis(newIndex sc.U32) sc.Option[sc.Sequence[primitives.AccountId]] {
	return sc.Option[sc.Sequence[primitives.AccountId]]{}
}

func (dm DefaultManager) EndSession(index sc.U32)   {}
func (dm DefaultManager) StartSession(index sc.U32) {}

// ShouldEndSession decides whether the session should be ended.
type ShouldEndSession interface {
	ShouldEndSession(blockNumber sc.U64) bool
}

type PeriodicSessions struct {
	Period sc.U64
	Offset sc.U64
}

func NewPeriodicSessions(period sc.U64, offset sc.U64) PeriodicSessions {
	return PeriodicSessions{
		Period: period,
		Offset: offset,
	}
}

func (ps PeriodicSessions) ShouldEndSession(now sc.U64) bool {
	return now >= ps.Offset && (((now - ps.Offset) % ps.Period) == 0)
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
