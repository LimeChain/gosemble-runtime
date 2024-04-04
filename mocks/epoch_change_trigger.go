package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/mock"
)

type EpochChangeTrigger struct {
	mock.Mock
}

func (t *EpochChangeTrigger) Trigger(now sc.U64) {
	t.Called(now)
}
