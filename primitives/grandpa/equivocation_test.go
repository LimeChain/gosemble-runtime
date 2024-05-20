package grandpa

import (
	"bytes"
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	grandpafinality "github.com/LimeChain/gosemble/primitives/grandpafinality"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	accountId, _ = primitives.NewAccountId(constants.OneAddress.FixedSequence...)
	pubKey       = primitives.Ed25519PublicKey(accountId)
	hash         = primitives.H256{
		FixedSequence: sc.FixedSequence[sc.U8]{
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1,
		},
	}

	signatureEd25519 = primitives.NewSignatureEd25519(
		sc.NewFixedSequence[sc.U8](
			64,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1,
		)...,
	)

	targetHashAndNumber = grandpafinality.TargetHashAndNumber{
		TargetHash:   primitives.H256{FixedSequence: hash.FixedSequence},
		TargetNumber: sc.U64(20),
	}

	equivocation = grandpafinality.Equivocation{
		RoundNumber: sc.U64(1),
		Identity:    pubKey,
		First: grandpafinality.Tuple2HashSignature{
			TargetHashAndNumber: targetHashAndNumber,
			SignatureEd25519:    signatureEd25519,
		},
		Second: grandpafinality.Tuple2HashSignature{
			TargetHashAndNumber: targetHashAndNumber,
			SignatureEd25519:    signatureEd25519,
		},
	}
)

func Test_NewEquivocationPrevote(t *testing.T) {
	result := NewEquivocationPrevote(equivocation)

	assert.Equal(t, Equivocation{sc.NewVaryingData(Prevote, equivocation)}, result)
}

func Test_NewEquivocationPrecommit(t *testing.T) {
	result := NewEquivocationPrecommit(equivocation)

	assert.Equal(t, Equivocation{sc.NewVaryingData(Precommit, equivocation)}, result)
}

func Test_DecodeEquivocation_Prevote(t *testing.T) {
	prevote := NewEquivocationPrevote(equivocation)
	buffer := bytes.NewBuffer(prevote.Bytes())

	result, err := DecodeEquivocation(buffer)

	assert.NoError(t, err)
	assert.Equal(t, prevote, result)
}

func Test_DecodeEquivocation_Prevote_Fails(t *testing.T) {
	prevote := NewEquivocationPrevote(equivocation)
	prevoteBytes := prevote.Bytes()
	buffer := bytes.NewBuffer(prevoteBytes[:len(prevoteBytes)-1])

	_, err := DecodeEquivocation(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_DecodeEquivocation_Precommit(t *testing.T) {
	precommit := NewEquivocationPrecommit(equivocation)
	buffer := bytes.NewBuffer(precommit.Bytes())

	result, err := DecodeEquivocation(buffer)

	assert.NoError(t, err)
	assert.Equal(t, precommit, result)
}

func Test_DecodeEquivocation_Precommit_Fails(t *testing.T) {
	precommit := NewEquivocationPrecommit(equivocation)
	precommitBytes := precommit.Bytes()

	buffer := bytes.NewBuffer(precommitBytes[:len(precommitBytes)-1])

	_, err := DecodeEquivocation(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_DecodeEquivocation_Fails(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x0})

	_, err := DecodeEquivocation(buffer)

	assert.Equal(t, io.EOF, err)
}
