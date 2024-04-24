package grandpa

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/session"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	KeyType                primitives.PublicKeyType
	MaxAuthorities         sc.U32
	MaxNominators          sc.U32
	MaxSetIdSessionEntries sc.U64
	SystemModule           system.Module
	SessionModule          session.Module
}

func NewConfig(
	keyType primitives.PublicKeyType,
	maxAuthorities sc.U32,
	maxNominators sc.U32,
	maxSetIdSessionEntries sc.U64,
	systemModule system.Module,
	sessionModule session.Module,
) *Config {
	return &Config{
		KeyType:                keyType,
		MaxAuthorities:         maxAuthorities,
		MaxNominators:          maxNominators,
		MaxSetIdSessionEntries: maxSetIdSessionEntries,
		SystemModule:           systemModule,
		SessionModule:          sessionModule,
	}
}
