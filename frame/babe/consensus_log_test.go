package babe

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

func Test_NewConsensusLogNextEpochData(t *testing.T) {
	assert.Equal(t,
		ConsensusLog{sc.NewVaryingData(ConsensusLogNextEpochData, nextEpochDescriptor)},
		NewConsensusLogNextEpochData(nextEpochDescriptor),
	)
}

func Test_NewConsensusLogOnDisabled(t *testing.T) {
	assert.Equal(t,
		ConsensusLog{sc.NewVaryingData(ConsensusLogOnDisabled, authorityIndex)},
		NewConsensusLogOnDisabled(authorityIndex),
	)
}

func Test_NewConsensusLogNextConfigData(t *testing.T) {
	assert.Equal(t,
		ConsensusLog{sc.NewVaryingData(ConsensusLogNextConfigData, nextConfigDescriptor)},
		NewConsensusLogNextConfigData(nextConfigDescriptor),
	)
}
