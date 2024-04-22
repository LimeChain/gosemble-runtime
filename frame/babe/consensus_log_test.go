package babe

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

func Test_NewNextEpochDataConsensusLog(t *testing.T) {
	assert.Equal(t,
		ConsensusLog{sc.NewVaryingData(NextEpochData, nextEpochDescriptor)},
		NewNextEpochDataConsensusLog(nextEpochDescriptor),
	)
}

func Test_NewOnDisabledConsensusLog(t *testing.T) {
	assert.Equal(t,
		ConsensusLog{sc.NewVaryingData(OnDisabled, authorityIndex)},
		NewOnDisabledConsensusLog(authorityIndex),
	)
}

func Test_NewNextConfigDataConsensusLog(t *testing.T) {
	assert.Equal(t,
		ConsensusLog{sc.NewVaryingData(NextConfigData, nextConfigDescriptor)},
		NewNextConfigDataConsensusLog(nextConfigDescriptor),
	)
}
