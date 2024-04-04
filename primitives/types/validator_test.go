package types

import (
	"bytes"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

var (
	expectBytesValidator, _ = hex.DecodeString("01000000000000000001000000000000000000010000000000000000000100000101000101000001010100000000000000000001000000000000000000010000")
)

var (
	targetValidator = Validator{
		AccountId:   accountId,
		AuthorityId: targetSr25519PublicKey,
	}
)

func Test_Validator_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetValidator.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, expectBytesValidator, buffer.Bytes())
}

func Test_DecodeValidator(t *testing.T) {
	buffer := bytes.NewBuffer(expectBytesValidator)

	result, err := DecodeValidator(buffer)
	assert.NoError(t, err)

	assert.Equal(t, targetValidator, result)
}

func Test_DecodeValidator_EOF(t *testing.T) {
	result, err := DecodeValidator(&bytes.Buffer{})
	assert.Error(t, io.EOF, err)

	assert.Equal(t, Validator{}, result)
}

func Test_DecodeValidator_PublicKey_EOF(t *testing.T) {
	buffer := bytes.NewBuffer(accountId.Bytes())
	result, err := DecodeValidator(buffer)
	assert.Error(t, io.EOF, err)

	assert.Equal(t, Validator{}, result)
}

func Test_Validator_Bytes(t *testing.T) {
	assert.Equal(t, expectBytesValidator, targetValidator.Bytes())
}
