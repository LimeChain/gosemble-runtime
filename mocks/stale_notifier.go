package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/mock"
)

type StaleNotifier struct {
	mock.Mock
}

func (sn *StaleNotifier) OnStalled(furtherWait sc.U64, median sc.U64) {
	sn.Called(furtherWait, median)
}
