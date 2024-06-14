package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/parachain"
	"github.com/stretchr/testify/mock"
)

type RelayChainStateProof struct {
	mock.Mock
}

func (m *RelayChainStateProof) ReadSlot() (sc.U64, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(sc.U64), nil
	}

	return args.Get(0).(sc.U64), args.Error(1).(error)
}

func (m *RelayChainStateProof) ReadUpgradeGoAheadSignal() (sc.Option[sc.U8], error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.U8]), nil
	}

	return args.Get(0).(sc.Option[sc.U8]), args.Error(1).(error)
}

func (m *RelayChainStateProof) ReadRestrictionSignal() (sc.Option[sc.U8], error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.U8]), nil
	}
	return args.Get(0).(sc.Option[sc.U8]), args.Error(1).(error)
}

func (m *RelayChainStateProof) ReadAbridgedHostConfiguration() (parachain.AbridgedHostConfiguration, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(parachain.AbridgedHostConfiguration), nil
	}
	return args.Get(0).(parachain.AbridgedHostConfiguration), args.Error(1).(error)
}

func (m *RelayChainStateProof) ReadIncludedParaHeadHash() sc.Option[sc.FixedSequence[sc.U8]] {
	args := m.Called()

	return args.Get(0).(sc.Option[sc.FixedSequence[sc.U8]])
}

func (m *RelayChainStateProof) ReadMessagingStateSnapshot(ahc parachain.AbridgedHostConfiguration) (parachain.MessagingStateSnapshot, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(parachain.MessagingStateSnapshot), nil
	}

	return args.Get(0).(parachain.MessagingStateSnapshot), args.Error(1).(error)
}
