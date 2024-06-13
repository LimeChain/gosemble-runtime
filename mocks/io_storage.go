package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/mock"
)

type IoStorage struct {
	mock.Mock
}

func (m *IoStorage) Append(key []byte, value []byte) {
	m.Called(key, value)
}

func (m *IoStorage) Clear(key []byte) {
	m.Called(key)
}

func (m *IoStorage) ClearPrefix(key []byte, limit []byte) {
	m.Called(key, limit)
}

func (m *IoStorage) Exists(key []byte) bool {
	args := m.Called(key)

	return args.Get(0).(bool)
}

func (m *IoStorage) Get(key []byte) (sc.Option[sc.Sequence[sc.U8]], error) {
	args := m.Called(key)
	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), nil
	}
	return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), args.Get(1).(error)
}

func (m *IoStorage) NextKey(key []byte) (sc.Option[sc.Sequence[sc.U8]], error) {
	args := m.Called(key)

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), args.Get(1).(error)
	}

	return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), nil
}

func (m *IoStorage) Read(key []byte, valueOut []byte, offset int32) (sc.Option[sc.U32], error) {
	args := m.Called(key, valueOut, offset)
	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.U32]), nil
	}

	return args.Get(0).(sc.Option[sc.U32]), args.Get(1).(error)
}

func (m *IoStorage) Root(version int32) []byte {
	args := m.Called(version)

	return args.Get(0).([]byte)
}

func (m *IoStorage) Set(key []byte, value []byte) {
	m.Called(key, value)
}
