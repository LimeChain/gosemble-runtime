package session

import sc "github.com/LimeChain/goscale"

// ShouldEndSession decides whether the session should be ended.
type ShouldEndSession interface {
	ShouldEndSession(blockNumber sc.U64) bool
}

type PeriodicSessions struct {
	Period sc.U64
	Offset sc.U64
}

func NewPeriodicSessions(period sc.U64, offset sc.U64) PeriodicSessions {
	return PeriodicSessions{
		Period: period,
		Offset: offset,
	}
}

func (ps PeriodicSessions) ShouldEndSession(now sc.U64) bool {
	return now >= ps.Offset && (((now - ps.Offset) % ps.Period) == 0)
}
