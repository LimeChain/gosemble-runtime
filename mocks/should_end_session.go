package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/mock"
)

type ShouldEndSession struct {
	mock.Mock
}

func (m *ShouldEndSession) ShouldEndSession(blockNumber sc.U64) bool {
	args := m.Called(blockNumber)

	return args.Get(0).(bool)
}
