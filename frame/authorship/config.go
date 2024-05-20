package authorship

import (
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	FindAuthor   primitives.FindAuthor[primitives.AccountId]
	EventHandler EventHandler
	SystemModule system.Module
}

func NewConfig(findAuthor primitives.FindAuthor[primitives.AccountId], eventHandler EventHandler, systemModule system.Module) *Config {
	return &Config{
		FindAuthor:   findAuthor,
		EventHandler: eventHandler,
		SystemModule: systemModule,
	}
}
