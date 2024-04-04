package types

import (
	"bytes"
	"errors"
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	bytesAddress32 = []byte{1, 1, 0, 1, 1, 0, 0, 1, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0}
)

func Test_NewAccountId(t *testing.T) {
	target := sc.BytesToFixedSequenceU8(bytesAddress32)
	expect := AccountId{
		FixedSequence: target,
	}

	result, err := NewAccountId(target...)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

func Test_NewAccountId_Fails(t *testing.T) {
	result, err := NewAccountId(5, 6)

	assert.Equal(t, errors.New("Address32 should be of size 32"), err)
	assert.Equal(t, AccountId{}, result)
}

func Test_AccountId_Encode(t *testing.T) {
	target, err := NewAccountId(sc.BytesToSequenceU8(bytesAddress32)...)
	assert.Nil(t, err)
	buffer := &bytes.Buffer{}

	err = target.Encode(buffer)
	assert.Nil(t, err)

	assert.Equal(t, bytesAddress32, buffer.Bytes())
}

func Test_AccountId_Bytes(t *testing.T) {
	target, err := NewAccountId(sc.BytesToSequenceU8(bytesAddress32)...)
	assert.Nil(t, err)

	result := target.Bytes()

	assert.Equal(t, bytesAddress32, result)
}

func Test_DecodeAccountId(t *testing.T) {
	expect, err := NewAccountId(sc.BytesToSequenceU8(bytesAddress32)...)
	assert.Nil(t, err)
	buffer := bytes.NewBuffer(bytesAddress32)

	result, err := DecodeAccountId(buffer)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

func Test_DecodeAccountId_Fails(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{5, 6})

	result, err := DecodeAccountId(buffer)

	assert.Equal(t, io.EOF, err)
	assert.Equal(t, AccountId{}, result)
}

func Test_DecodeSequenceAccountId(t *testing.T) {
	accId, err := NewAccountId(sc.BytesToSequenceU8(bytesAddress32)...)
	assert.Nil(t, err)

	expect := sc.Sequence[AccountId]{accId}

	buffer := &bytes.Buffer{}
	buffer.WriteByte(4)
	buffer.Write(bytesAddress32)

	result, err := DecodeSequenceAccountId(buffer)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}
