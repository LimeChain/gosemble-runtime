package types

import (
	"bytes"
	"math/big"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedIsNewLogic, _ = new(big.Int).SetString("80000000000000000000000000000000", 16)
	expectedExtraFlags    = ExtraFlags{sc.NewU128(expectedIsNewLogic)}
)

func TestExtraFlags(t *testing.T) {
	// Test DefaultExtraFlags
	ef := DefaultExtraFlags()
	assert.Equal(t, expectedExtraFlags, ef)
	assert.Equal(t, expectedIsNewLogic, ef.ToBigInt())

	// Test Encode
	buffer := &bytes.Buffer{}
	err := ef.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedExtraFlags.Bytes(), buffer.Bytes())

	// Test Decode
	decodedEf, err := DecodeExtraFlags(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedExtraFlags, decodedEf)
	assert.Equal(t, expectedIsNewLogic, decodedEf.ToBigInt())
	_, err = DecodeExtraFlags(buffer)
	assert.Error(t, err)

	// Test IsNewLogic
	newEf := ExtraFlags{}
	assert.False(t, newEf.IsNewLogic())
	// Test OldLogic
	assert.Equal(t, ExtraFlags{}, newEf.OldLogic())
	// Test SetNewLogic
	newEf = newEf.SetNewLogic()
	assert.True(t, newEf.IsNewLogic())
	assert.Equal(t, expectedExtraFlags, newEf)
	assert.Equal(t, expectedExtraFlags, newEf.OldLogic())
	assert.Equal(t, expectedIsNewLogic, newEf.ToBigInt())
	// assert.Equal(t, expectedBytes, buffer.Bytes())

	// // Test Bytes
	// assert.Equal(t, expectedBytes, ef.ToBigInt().Bytes())

	// // Test OldLogic
	// oldEf := ef.OldLogic()
	// assert.Equal(t, ef, oldEf)

	// // Assert IsNewLogic is false before setting new logic
	// // assert.False(t, ef.IsNewLogic())

	// // Test SetNewLogic
	// newEf := ef.SetNewLogic()
	// assert.Equal(t, expected, newEf.U128.ToBigInt())

	// // Test IsNewLogic
	// assert.True(t, newEf.IsNewLogic())

	// // Test DecodeExtraFlags
	// decodedEf, err := DecodeExtraFlags(buffer)
	// assert.NoError(t, err)
	// assert.Equal(t, newEf, decodedEf)
}
