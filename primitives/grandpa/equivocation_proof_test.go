package grandpa

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	grandpafinality "github.com/LimeChain/gosemble/primitives/grandpafinality"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	accountId0, _ = primitives.NewAccountId(constants.ZeroAddress.FixedSequence...)
	accountId1, _ = primitives.NewAccountId(constants.OneAddress.FixedSequence...)

	prevoteEquivocation = grandpafinality.Equivocation{
		RoundNumber: sc.U64(3),
		Identity:    primitives.Ed25519PublicKey(accountId1),
		First: grandpafinality.Tuple2HashSignature{
			TargetHashAndNumber: targetHashAndNumber,
			SignatureEd25519:    signatureEd25519,
		},
		Second: grandpafinality.Tuple2HashSignature{
			TargetHashAndNumber: targetHashAndNumber,
			SignatureEd25519:    signatureEd25519,
		},
	}

	precommitEquivocation = grandpafinality.Equivocation{
		RoundNumber: sc.U64(4),
		Identity:    primitives.Ed25519PublicKey(accountId0),
		First: grandpafinality.Tuple2HashSignature{
			TargetHashAndNumber: targetHashAndNumber,
			SignatureEd25519:    signatureEd25519,
		},
		Second: grandpafinality.Tuple2HashSignature{
			TargetHashAndNumber: targetHashAndNumber,
			SignatureEd25519:    signatureEd25519,
		},
	}

	equivocationProofPrevote = EquivocationProof{
		SetId:        sc.U64(2),
		Equivocation: NewEquivocationPrevote(prevoteEquivocation),
	}

	equivocationProofPrecommit = EquivocationProof{
		SetId:        sc.U64(3),
		Equivocation: NewEquivocationPrecommit(precommitEquivocation),
	}
)

var (
	equivocationProofPrevoteBytes, _ = hex.DecodeString("0200000000000000000300000000000000000000000000000000000000000000000000000000000000000000000000000101010101010101010101010101010101010101010101010101010101010101011400000000000000010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101140000000000000001010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101")
)

func Test_EquivocationProof_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := equivocationProofPrevote.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, equivocationProofPrevoteBytes, buffer.Bytes())
}

func Test_EquivocationProof_Bytes(t *testing.T) {
	assert.Equal(t, equivocationProofPrevoteBytes, equivocationProofPrevote.Bytes())
}

func Test_DecodeEquivocationProof(t *testing.T) {
	buffer := bytes.NewBuffer(equivocationProofPrevoteBytes)

	result, err := DecodeEquivocationProof(buffer)

	assert.NoError(t, err)
	assert.Equal(t, equivocationProofPrevote, result)
}

func Test_DecodeEquivocationProof_Fails_To_Decode_SetId(t *testing.T) {
	buffer := bytes.NewBuffer(equivocationProofPrevoteBytes[:1])

	_, err := DecodeEquivocationProof(buffer)

	assert.Error(t, err)
}

func Test_DecodeEquivocationProof_Fails_To_Decode_Equivocation(t *testing.T) {
	buffer := bytes.NewBuffer(equivocationProofPrevoteBytes[:len(equivocationProofPrevoteBytes)-1])

	_, err := DecodeEquivocationProof(buffer)

	assert.Error(t, err)
}

func Test_EquivocationProof_Prevote_Round(t *testing.T) {
	result, err := equivocationProofPrevote.Round()

	assert.NoError(t, err)
	assert.Equal(t, sc.U64(3), result)
}

func Test_EquivocationProof_Precommit_Round(t *testing.T) {
	result, err := equivocationProofPrecommit.Round()

	assert.NoError(t, err)
	assert.Equal(t, sc.U64(4), result)
}

func Test_EquivocationProof_Prevote_Offender(t *testing.T) {
	result, err := equivocationProofPrevote.Offender()

	assert.NoError(t, err)
	assert.Equal(t, accountId1, result)
}

func Test_EquivocationProof_Precommit_Offender(t *testing.T) {
	result, err := equivocationProofPrecommit.Offender()

	assert.NoError(t, err)
	assert.Equal(t, accountId0, result)
}
