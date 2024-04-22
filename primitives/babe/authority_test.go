package babe

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	authority = Authority{
		Key:    types.Sr25519PublicKey{FixedSequence: sc.BytesToFixedSequenceU8(make([]byte, 32))},
		Weight: sc.U64(2),
	}
)

var (
	authorityBytes = []byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		2, 0, 0, 0, 0, 0, 0, 0,
	}
)

func Test_Authority_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := authority.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, authorityBytes, buffer.Bytes())
}

func Test_Authority_Bytes(t *testing.T) {
	assert.Equal(t, authorityBytes, authority.Bytes())
}

func Test_DecodeAuthority(t *testing.T) {
	buffer := bytes.NewBuffer(authorityBytes)

	result, err := DecodeAuthority(buffer)

	assert.NoError(t, err)
	assert.Equal(t, authority, result)
}
