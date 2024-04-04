package aura

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	KeyType                    primitives.PublicKeyType
	DbWeight                   primitives.RuntimeDbWeight
	MinimumPeriod              sc.U64
	MaxAuthorities             sc.U32
	AllowMultipleBlocksPerSlot bool
	SystemDigest               func() (primitives.Digest, error)
	LogDepositor               system.LogDepositor
	DisabledValidators         primitives.DisabledValidators
}

func NewConfig(keyType primitives.PublicKeyType, dbWeight primitives.RuntimeDbWeight, minimumPeriod sc.U64, maxAuthorities sc.U32, allowMultipleBlocksPerSlot bool, systemDigest func() (primitives.Digest, error), logDepositor system.LogDepositor, disabledValidators primitives.DisabledValidators) *Config {
	return &Config{
		KeyType:                    keyType,
		DbWeight:                   dbWeight,
		MinimumPeriod:              minimumPeriod,
		MaxAuthorities:             maxAuthorities,
		AllowMultipleBlocksPerSlot: allowMultipleBlocksPerSlot,
		SystemDigest:               systemDigest,
		LogDepositor:               logDepositor,
		DisabledValidators:         disabledValidators,
	}
}
