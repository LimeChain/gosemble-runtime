package types

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedAccountDataBytes, _ = hex.DecodeString(
		"01000000000000000000000000000000020000000000000000000000000000000300000000000000000000000000000004000000000000000000000000000000",
	)
)

var (
	targetAccountData = AccountData{
		Free:     sc.NewU128(1),
		Reserved: sc.NewU128(2),
		Frozen:   sc.NewU128(3),
		Flags:    ExtraFlags{sc.NewU128(4)},
	}
)

func Test_AccountData_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetAccountData.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccountDataBytes, buffer.Bytes())
}

func Test_AccountData_Bytes(t *testing.T) {
	assert.Equal(t, expectedAccountDataBytes, targetAccountData.Bytes())
}

func Test_DecodeAccountData(t *testing.T) {
	buffer := bytes.NewBuffer(expectedAccountDataBytes)

	result, err := DecodeAccountData(buffer)
	assert.NoError(t, err)

	assert.Equal(t, targetAccountData, result)
}

func Test_DecodeAccountData_EOF(t *testing.T) {
	examples := []struct {
		label string
		input string
	}{
		{"empty", ""},
		{"only free", "01000000000000000000000000000000"},
		{"free and reserved, no misc and fee", "0100000000000000000000000000000002000000000000000000000000000000"},
		{"free, reserved and misc, no fee", "010000000000000000000000000000000200000000000000000000000000000003000000000000000000000000000000"},
	}

	for _, example := range examples {
		t.Run(example.label, func(t *testing.T) {
			b, err := hex.DecodeString(example.input)
			assert.Nil(t, err)

			result, err := DecodeAccountData(bytes.NewBuffer(b))

			assert.Equal(t, err, io.EOF)
			assert.Equal(t, AccountData{}, result)
		})
	}
}

func Test_Total(t *testing.T) {
	assert.Equal(t, sc.NewU128(3), targetAccountData.Total())
}
