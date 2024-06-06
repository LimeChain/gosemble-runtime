package grandpa

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/session"
	"github.com/LimeChain/gosemble/frame/system"
	staking "github.com/LimeChain/gosemble/primitives/staking"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	DbWeight                 types.RuntimeDbWeight
	KeyType                  primitives.PublicKeyType
	MaxAuthorities           sc.U32
	MaxNominators            sc.U32
	MaxSetIdSessionEntries   sc.U64
	KeyOwnerProof            KeyOwnerProofSystem
	EquivocationReportSystem staking.OffenceReportSystem
	SystemModule             system.Module
	SessionModule            session.Module
}

func NewConfig(
	dbWeight types.RuntimeDbWeight,
	keyType primitives.PublicKeyType,
	maxAuthorities sc.U32,
	maxNominators sc.U32,
	maxSetIdSessionEntries sc.U64,
	keyOwnerProof KeyOwnerProofSystem,
	equivocationReportSystem staking.OffenceReportSystem,
	systemModule system.Module,
	sessionModule session.Module,
) *Config {
	return &Config{
		DbWeight:                 dbWeight,
		KeyType:                  keyType,
		MaxAuthorities:           maxAuthorities,
		MaxNominators:            maxNominators,
		MaxSetIdSessionEntries:   maxSetIdSessionEntries,
		KeyOwnerProof:            keyOwnerProof,
		EquivocationReportSystem: equivocationReportSystem,
		SystemModule:             systemModule,
		SessionModule:            sessionModule,
	}
}
