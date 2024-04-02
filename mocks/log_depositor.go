package mocks

import (
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type LogDepositor struct {
	mock.Mock
}

func (m *LogDepositor) DepositLog(item types.DigestItem) {
	m.Called(item)
}
