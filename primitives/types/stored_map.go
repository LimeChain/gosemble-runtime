package types

import (
	sc "github.com/LimeChain/goscale"
)

type StoredMap interface {
	EventDepositor
	Get(key AccountId) (AccountInfo, error)
	CanDecProviders(who AccountId) (bool, error)
	DecConsumers(who AccountId) error
	DecProviders(who AccountId) (DecRefStatus, error)
	IncConsumers(who AccountId) error
	IncConsumersWithoutLimit(who AccountId) error
	IncProviders(who AccountId) (IncRefStatus, error)
	Insert(who AccountId, data AccountData) (sc.Encodable, error)
	TryMutateExists(who AccountId, f func(who *AccountData) (sc.Encodable, error)) (sc.Encodable, error)
}
