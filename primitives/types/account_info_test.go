package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedAccountInfoBytes, _ = hex.DecodeString(
		"0100000002000000030000000400000005000000000000000000000000000000060000000000000000000000000000000700000000000000000000000000000000000000000000000000000000000080",
	)
)
var (
	targetAccountInfo = AccountInfo{
		Nonce:       AccountIndex(1),
		Consumers:   RefCount(2),
		Providers:   RefCount(3),
		Sufficients: RefCount(4),
		Data: AccountData{
			Free:     sc.NewU128(5),
			Reserved: sc.NewU128(6),
			Frozen:   sc.NewU128(7),
			Flags:    DefaultExtraFlags(),
		},
	}
)

func Test_AccountInfo_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetAccountInfo.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccountInfoBytes, buffer.Bytes())
}

func Test_AccountInfo_Bytes(t *testing.T) {
	assert.Equal(t, expectedAccountInfoBytes, targetAccountInfo.Bytes())
}

func Test_DecodeAccountInfo(t *testing.T) {
	buffer := bytes.NewBuffer(expectedAccountInfoBytes)

	result, err := DecodeAccountInfo(buffer)
	assert.NoError(t, err)

	assert.Equal(t, targetAccountInfo, result)
}

// func Test_AccountInfo_Frozen(t *testing.T) {
// 	assert.Equal(t, sc.NewU128(8), targetAccountInfo.Frozen(ReasonsAll))
// 	assert.Equal(t, sc.NewU128(7), targetAccountInfo.Frozen(ReasonsMisc))
// 	assert.Equal(t, sc.NewU128(7), targetAccountInfo.Frozen(ReasonsFee))
// 	assert.Equal(t, sc.NewU128(0), targetAccountInfo.Frozen(3))
// }

// func Test_AccountInfo_Frozen_WithGreaterMiscFrozen(t *testing.T) {
// 	targetAccountInfo = AccountInfo{}
// 	targetAccountInfo.Data.Frozen = sc.NewU128(8)

// 	assert.Equal(t, sc.NewU128(9), targetAccountInfo.Frozen(ReasonsAll))
// }
