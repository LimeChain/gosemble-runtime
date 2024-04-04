package session

import (
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
