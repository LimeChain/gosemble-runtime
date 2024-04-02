package session

import (
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	DbWeight     types.RuntimeDbWeight
	BlockWeights types.BlockWeights
	Module       system.Module
	SessionEnder ShouldEndSession
	Handler      Handler
	Manager      Manager
}

func NewConfig(dbWeight types.RuntimeDbWeight, blockWeights types.BlockWeights, module system.Module, sessionEnder ShouldEndSession, handler Handler, manager Manager) Config {
	return Config{
		dbWeight,
		blockWeights,
		module,
		sessionEnder,
		handler,
		manager,
	}
}
