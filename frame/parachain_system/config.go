package parachain_system

import (
	"github.com/LimeChain/gosemble/frame/parachain_info"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage                    io.Storage
	DbWeight                   primitives.RuntimeDbWeight
	CheckAssociatedRelayNumber CheckAssociatedRelayNumber
	SelfParaId                 parachain_info.Module
	systemModule               system.Module
	ConsensusHook              ConsensusHook
}

func NewConfig(storage io.Storage, dbWeight primitives.RuntimeDbWeight, checkAssociatedRelayNumber CheckAssociatedRelayNumber, selfParaId parachain_info.Module, systemModule system.Module, consensusHook ConsensusHook) Config {
	return Config{
		Storage:                    storage,
		DbWeight:                   dbWeight,
		CheckAssociatedRelayNumber: checkAssociatedRelayNumber,
		SelfParaId:                 selfParaId,
		systemModule:               systemModule,
		ConsensusHook:              consensusHook,
	}
}
