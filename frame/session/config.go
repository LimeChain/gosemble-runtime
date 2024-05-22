package session

import (
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage      io.Storage
	DbWeight     types.RuntimeDbWeight
	BlockWeights types.BlockWeights
	Module       system.Module
	SessionEnder ShouldEndSession
	Handler      Handler
	Manager      Manager
}

func NewConfig(storage io.Storage, dbWeight types.RuntimeDbWeight, blockWeights types.BlockWeights, module system.Module, sessionEnder ShouldEndSession, handler Handler, manager Manager) Config {
	return Config{
		storage,
		dbWeight,
		blockWeights,
		module,
		sessionEnder,
		handler,
		manager,
	}
}
