package mocks

import (
	"github.com/ChainSafe/gossamer/lib/runtime/storage"
	"github.com/stretchr/testify/mock"
)

type HostEnvironment struct {
	mock.Mock
	IoStorage
	IoTransactionBroker
}

func (m *HostEnvironment) SetTrieState(state *storage.TrieState) {
	m.Called(state)
}
