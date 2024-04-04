package babe

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/session"
	"github.com/LimeChain/gosemble/frame/system"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	KeyType            primitives.PublicKeyType
	EpochConfig        babetypes.EpochConfiguration
	EpochDuration      sc.U64
	EpochChangeTrigger EpochChangeTrigger
	SessionModule      session.Module
	MaxAuthorities     sc.U32
	MinimumPeriod      sc.U64
	SystemDigest       func() (primitives.Digest, error)
	SystemModule       system.Module
}

func NewConfig(
	keyType primitives.PublicKeyType,
	epochConfig babetypes.EpochConfiguration,
	epochDuration sc.U64,
	epochChangeTrigger EpochChangeTrigger,
	sessionModule session.Module,
	maxAuthorities sc.U32,
	minimumPeriod sc.U64,
	systemDigest func() (primitives.Digest, error),
	systemModule system.Module,
) *Config {
	return &Config{
		KeyType:            keyType,
		EpochConfig:        epochConfig,
		EpochDuration:      epochDuration,
		EpochChangeTrigger: epochChangeTrigger,
		SessionModule:      sessionModule,
		MaxAuthorities:     maxAuthorities,
		MinimumPeriod:      minimumPeriod,
		SystemDigest:       systemDigest,
		SystemModule:       systemModule,
	}
}
