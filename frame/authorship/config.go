package authorship

import (
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage      io.Storage
	FindAuthor   primitives.FindAuthor[primitives.AccountId]
	EventHandler EventHandler
	SystemModule system.Module
}

func NewConfig(storage io.Storage, findAuthor primitives.FindAuthor[primitives.AccountId], eventHandler EventHandler, systemModule system.Module) *Config {
	return &Config{
		Storage:      storage,
		FindAuthor:   findAuthor,
		EventHandler: eventHandler,
		SystemModule: systemModule,
	}
}
