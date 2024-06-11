package session

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

// Proof of membership of a specific key in a given session.
type MembershipProof struct {
	// The session index on which the specific key is a member.
	SessionIndex sc.U32
	// Trie nodes of a merkle proof of session membership.
	TrieNodes sc.Sequence[sc.Sequence[sc.U8]]
	// The validator count of the session on which the specific key is a member.
	Validators sc.U32
}

func (m MembershipProof) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		m.SessionIndex,
		m.TrieNodes,
		m.Validators,
	)
}

func (m MembershipProof) Bytes() []byte {
	return sc.EncodedBytes(m)
}

func DecodeMembershipProof(buffer *bytes.Buffer) (MembershipProof, error) {
	sessionIndex, err := sc.DecodeU32(buffer)
	if err != nil {
		return MembershipProof{}, err
	}

	trieNodes, err := sc.DecodeSequenceWith(buffer, func(buffer *bytes.Buffer) (sc.Sequence[sc.U8], error) {
		return sc.DecodeSequence[sc.U8](buffer)
	})
	if err != nil {
		return MembershipProof{}, err
	}

	validatorCount, err := sc.DecodeU32(buffer)
	if err != nil {
		return MembershipProof{}, err
	}

	return MembershipProof{
		SessionIndex: sessionIndex,
		TrieNodes:    trieNodes,
		Validators:   validatorCount,
	}, nil
}

func (m MembershipProof) Session() sc.U32 {
	return m.SessionIndex
}

func (m MembershipProof) ValidatorCount() sc.U32 {
	return m.Validators
}
