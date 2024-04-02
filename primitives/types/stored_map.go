package types

import (
	sc "github.com/LimeChain/goscale"
)

type StoredMap interface {
	EventDepositor
	Get(key AccountId) (AccountInfo, error)
	CanDecProviders(who AccountId) (bool, error)
	TryMutateExists(who AccountId, f func(who *AccountData) (sc.Encodable, error)) (sc.Encodable, error)
	TryMutateExistsNew(who AccountId, f func(who *AccountData) (sc.Encodable, error)) (sc.Encodable, error)
	TryMutateExistsNoClosure(who AccountId, data AccountData) error
	IncProviders(who AccountId) (IncRefStatus, error)
	DecProviders(who AccountId) (DecRefStatus, error)
	IncConsumers(who AccountId) error
	DecConsumers(who AccountId) error
	IncConsumersWithoutLimit(who AccountId) error
}
