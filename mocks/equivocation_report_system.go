package mocks

import (
	sc "github.com/LimeChain/goscale"
	grandpatypes "github.com/LimeChain/gosemble/primitives/grandpa"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type EquivocationReportSystem struct {
	mock.Mock
}

func (m *EquivocationReportSystem) PublishEvidence(equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof) error {
	args := m.Called(equivocationProof, keyOwnerProof)
	return args.Error(0)
}

func (m *EquivocationReportSystem) ProcessEvidence(reporter sc.Option[primitives.AccountId], equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof) error {
	args := m.Called(reporter, equivocationProof, keyOwnerProof)
	return args.Error(0)
}
