package io

import (
	"github.com/LimeChain/gosemble/env"
	"github.com/LimeChain/gosemble/utils"
)

type Offchain interface {
	SubmitTransaction(value []byte) []byte
}

type offchain struct {
	memoryTranslator utils.WasmMemoryTranslator
}

func NewOffchain() Offchain {
	return offchain{
		memoryTranslator: utils.NewMemoryTranslator(),
	}
}

func (o offchain) SubmitTransaction(value []byte) []byte {
	valueOffsetSize := o.memoryTranslator.BytesToOffsetAndSize(value)
	resOffsetSize := env.ExtOffchainSubmitTransactionVersion1(valueOffsetSize)
	offset, size := o.memoryTranslator.Int64ToOffsetAndSize(resOffsetSize)
	return o.memoryTranslator.GetWasmMemorySlice(offset, size)
}
