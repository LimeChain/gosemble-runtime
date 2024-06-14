package mocks

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/parachain"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type RuntimeDecoder struct {
	mock.Mock
}

func (m *RuntimeDecoder) DecodeBlock(buffer *bytes.Buffer) (types.Block, error) {
	args := m.Called(buffer)

	if args.Get(1) == nil {
		return args.Get(0).(types.Block), nil
	}

	return args.Get(0).(types.Block), args.Get(1).(error)
}

func (m *RuntimeDecoder) DecodeParachainBlockData(blockData sc.Sequence[sc.U8]) (parachain.BlockData, error) {
	args := m.Called(blockData)

	if args.Get(1) == nil {
		return args.Get(0).(parachain.BlockData), nil
	}

	return args.Get(0).(parachain.BlockData), args.Get(1).(error)
}

func (m *RuntimeDecoder) DecodeUncheckedExtrinsic(buffer *bytes.Buffer) (types.UncheckedExtrinsic, error) {
	args := m.Called(buffer)

	if args.Get(1) == nil {
		return args.Get(0).(types.UncheckedExtrinsic), nil
	}

	return args.Get(0).(types.UncheckedExtrinsic), args.Get(1).(error)
}

func (m *RuntimeDecoder) DecodeCall(buffer *bytes.Buffer) (primitives.Call, error) {
	args := m.Called(buffer)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.Call), nil
	}

	return args.Get(0).(primitives.Call), args.Get(1).(error)
}
