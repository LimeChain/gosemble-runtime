package babe

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	KeyType        primitives.PublicKeyType
	EpochConfig    BabeEpochConfiguration
	EpochDuration  sc.U64
	MinimumPeriod  sc.U64
	MaxAuthorities sc.U32
}

func NewConfig(keyType primitives.PublicKeyType, epochConfig BabeEpochConfiguration, epochDuration sc.U64, minimumPeriod sc.U64, maxAuthorities sc.U32) *Config {
	return &Config{
		KeyType:        keyType,
		EpochConfig:    epochConfig,
		EpochDuration:  epochDuration,
		MinimumPeriod:  minimumPeriod,
		MaxAuthorities: maxAuthorities,
	}
}
