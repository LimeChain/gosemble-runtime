package pvf

import (
	"bytes"
	"fmt"
	"math"
	"math/big"

	"github.com/ChainSafe/gossamer/lib/runtime/storage"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/log"
)

type HostEnvironment struct {
	trieState *storage.TrieState
	logger    log.RuntimeLogger
}

func NewHostEnvironment(logger log.RuntimeLogger) *HostEnvironment {
	return &HostEnvironment{
		logger:    logger,
		trieState: nil,
	}
}

func (he *HostEnvironment) SetTrieState(state *storage.TrieState) {
	he.trieState = state
}

func (he HostEnvironment) Append(key []byte, value []byte) {
	cp := make([]byte, len(value))
	copy(cp, value)

	err := he.storageAppend(key, cp)
	if err != nil {
		he.logger.Warnf("failed appending to storage: %s", err)
	}
}

func (he HostEnvironment) Clear(key []byte) {
	err := he.trieState.Delete(key)
	if err != nil {
		he.logger.Critical(err.Error())
	}
}

func (he HostEnvironment) ClearPrefix(prefix []byte, limitBytes []byte) {
	he.logger.Debugf("prefix: 0x%x", prefix)

	limitOption, err := sc.DecodeOption[sc.U32](bytes.NewBuffer(limitBytes))
	if err != nil {
		he.logger.Criticalf("failed scale decoding limit: %s", err)
	}

	var limitPtr uint32
	if limitOption.HasValue {
		value := uint32(limitOption.Value)
		limitPtr = value
	} else {
		maxLimit := uint32(math.MaxUint32)
		limitPtr = maxLimit
	}

	numRemoved, all, err := he.trieState.ClearPrefixLimit(prefix, limitPtr)
	if err != nil {
		he.logger.Criticalf("failed to clear prefix limit: %s", err)
	}

	_ = toKillStorageResultEnum(all, numRemoved)
	// TODO: func signature is not valid
}

func (he HostEnvironment) Exists(key []byte) bool {
	he.logger.Debugf("key: 0x%x", key)

	value := he.trieState.Get(key)
	if value != nil {
		return true
	}

	return false
}

func (he HostEnvironment) Get(key []byte) (sc.Option[sc.Sequence[sc.U8]], error) {
	value := he.trieState.Get(key)
	he.logger.Debugf("value: 0x%x", value)

	if value == nil {
		return sc.NewOption[sc.Sequence[sc.U8]](nil), nil
	}

	return sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(value)), nil
}

func (he HostEnvironment) NextKey(key []byte) (sc.Option[sc.Sequence[sc.U8]], error) {
	next := he.trieState.NextKey(key)
	he.logger.Debugf(
		"key: 0x%x; next key 0x%x",
		key, next)

	if len(next) == 0 {
		sc.NewOption[sc.Sequence[sc.U8]](nil)
	}

	return sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(next)), nil
}

func (he HostEnvironment) Read(key []byte, valueOut []byte, offset int32) (sc.Option[sc.U32], error) {
	he.logger.Debugf(
		"READ key 0x%x has value 0x%x",
		key, valueOut)

	value := he.trieState.Get(key)
	if value == nil {
		return sc.NewOption[sc.U32](nil), nil
	}

	var data []byte
	switch {
	case uint32(offset) <= uint32(len(value)):
		data = value[offset:]
	default:
		data = value[len(value):]
	}

	if len(valueOut) >= len(data) {
		copy(valueOut, data)
	} else {
		copy(valueOut, data[0:len(valueOut)])
	}

	return sc.NewOption[sc.U32](sc.U32(len(data))), nil
}

func (he HostEnvironment) Root(version int32) []byte {
	root, err := he.trieState.Root()
	if err != nil {
		he.logger.Criticalf("failed to get storage root: %s", err)
	}

	return root[:]
}

func (he HostEnvironment) Set(key []byte, value []byte) {
	cp := make([]byte, len(value))
	copy(cp, value)

	he.logger.Debugf(
		"key 0x%x has value 0x%x",
		key, value)
	err := he.trieState.Put(key, cp)
	if err != nil {
		he.logger.Criticalf("failed to set value: key [%x], value [%x], err: %s", key, value, err)
	}
}

func (he HostEnvironment) Start() {
	he.trieState.StartTransaction()
}

func (he HostEnvironment) Commit() {
	he.trieState.CommitTransaction()
}

func (he HostEnvironment) Rollback() {
	he.trieState.RollbackTransaction()
}

func (he HostEnvironment) storageAppend(key, valueToAppend []byte) (err error) {
	// this function assumes the item in storage is a SCALE encoded array of items
	// the valueToAppend is a new item, so it appends the item and increases the length prefix by 1
	currentValue := he.trieState.Get(key)

	var value []byte
	if len(currentValue) == 0 {
		nextLength := 1
		encodedLength := sc.ToCompact(nextLength).Bytes()

		value = make([]byte, len(encodedLength)+len(valueToAppend))
		// append new length prefix to start of items array
		copy(value, encodedLength)
		copy(value[len(encodedLength):], valueToAppend)
	} else {
		buffer := bytes.NewBuffer(currentValue)
		currentLength, err := sc.DecodeCompact[sc.U128](buffer)
		if err != nil {
			he.logger.Tracef(
				"item in storage is not SCALE encoded, overwriting at key 0x%x", key)
			value = make([]byte, 1+len(valueToAppend))
			value[0] = 4
			copy(value[1:], valueToAppend)
		} else {
			lengthBytes := currentLength.Bytes()

			// increase length by 1
			nextLength := big.NewInt(0).Add(currentLength.ToBigInt(), big.NewInt(1))
			nextLengthBytes := sc.Compact{Number: sc.NewU128(nextLength)}.Bytes()

			// append new item, pop off number of bytes required for length encoding,
			// since we're not using old scale.Decoder
			value = make([]byte, len(nextLengthBytes)+len(currentValue)-len(lengthBytes)+len(valueToAppend))
			// append new length prefix to start of items array
			i := 0
			copy(value[i:], nextLengthBytes)
			i += len(nextLengthBytes)
			copy(value[i:], currentValue[len(lengthBytes):])
			i += len(currentValue) - len(lengthBytes)
			copy(value[i:], valueToAppend)
		}
	}

	err = he.trieState.Put(key, value)
	if err != nil {
		return fmt.Errorf("putting key and value in storage: %w", err)
	}

	return nil
}

// toKillStorageResultEnum encodes the `allRemoved` flag and
// the `numRemoved` uint32 to a byte slice and returns it.
// The format used is:
// Byte 0: 1 if allRemoved is false, 0 otherwise
// Byte 1-5: scale encoding of numRemoved (up to 4 bytes)
func toKillStorageResultEnum(allRemoved bool, numRemoved uint32) []byte {
	encodedNumRemoved := sc.U32(numRemoved).Bytes()

	encodedEnumValue := make([]byte, len(encodedNumRemoved)+1)
	if !allRemoved {
		// At least one key resides in the child trie due to the supplied limit.
		encodedEnumValue[0] = 1
	}
	copy(encodedEnumValue[1:], encodedNumRemoved)

	return encodedEnumValue
}
