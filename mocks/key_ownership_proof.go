package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/session"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type KeyOwnerProofSystem struct {
	mock.Mock
}

func (k *KeyOwnerProofSystem) Prove(key [4]byte, authorityId primitives.AccountId) sc.Option[session.MembershipProof] {
	args := k.Called(key, authorityId)
	return args.Get(0).(sc.Option[session.MembershipProof])
}
