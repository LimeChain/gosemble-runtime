package mocks

import (
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type BlockExecutor struct {
	mock.Mock
}

func (m *BlockExecutor) ExecuteBlock(block types.Block) error {
	args := m.Called(block)
	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}
