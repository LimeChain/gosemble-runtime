package session

import (
	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_PeriodicSession_ShouldEndSession_True(t *testing.T) {
	target := NewPeriodicSessions(2, 2)
	now := sc.U64(10)

	assert.Equal(t, true, target.ShouldEndSession(now))
}

func Test_PeriodicSession_ShouldEndSession_False(t *testing.T) {
	target := NewPeriodicSessions(5, 1)
	now := sc.U64(10)

	assert.Equal(t, false, target.ShouldEndSession(now))
}
