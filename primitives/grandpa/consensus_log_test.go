package grandpa

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	median         = sc.U64(123)
	authorityIndex = sc.U64(10)
)

func Test_NewConsensusLogScheduledChange(t *testing.T) {
	result := NewConsensusLogScheduledChange(scheduledChange)

	assert.Equal(t, ConsensusLog{sc.NewVaryingData(ConsensusLogScheduledChange, scheduledChange)}, result)
}

func Test_NewConsensusLogForcedChange(t *testing.T) {
	result := NewConsensusLogForcedChange(median, scheduledChange)

	assert.Equal(t, ConsensusLog{sc.NewVaryingData(ConsensusLogForcedChange, median, scheduledChange)}, result)
}

func Test_NewConsensusLogOnDisabled(t *testing.T) {
	result := NewConsensusLogOnDisabled(authorityIndex)

	assert.Equal(t, ConsensusLog{sc.NewVaryingData(ConsensusLogOnDisabled, authorityIndex)}, result)
}

func Test_NewConsensusLogPause(t *testing.T) {
	result := NewConsensusLogPause(authorityIndex)

	assert.Equal(t, ConsensusLog{sc.NewVaryingData(ConsensusLogPause)}, result)
}

func Test_NewConsensusLogResume(t *testing.T) {
	result := NewConsensusLogResume(authorityIndex)

	assert.Equal(t, ConsensusLog{sc.NewVaryingData(ConsensusLogResume)}, result)
}
