package staking

import (
	sc "github.com/LimeChain/goscale"
	grandpatypes "github.com/LimeChain/gosemble/primitives/grandpa"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type OffenceReportSystem interface {
	PublishEvidence(equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof) error
	// CheckEvidence(_evidence: Evidence) TransactionValidityError
	ProcessEvidence(reporter sc.Option[primitives.AccountId], equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof) error
}

type DefaultOffenceReportSystem struct{}

func (d DefaultOffenceReportSystem) PublishEvidence(equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof) error {
	return nil
}

func (d DefaultOffenceReportSystem) ProcessEvidence(reporter sc.Option[primitives.AccountId], equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof) error {
	return nil
}
