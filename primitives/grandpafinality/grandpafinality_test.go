package grandpafinality

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	targetHashAndNumberBytes, _ = hex.DecodeString("01010101010101010101010101010101010101010101010101010101010101010a00000000000000")
	signatureEd25519Bytes, _    = hex.DecodeString("01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101")
	hashSignatureBytes, _       = hex.DecodeString("01010101010101010101010101010101010101010101010101010101010101010a0000000000000001010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101")
	equivocationBytes, _        = hex.DecodeString("0a00000000000000000000000000000000000000000000000000000000000000000000000000000101010101010101010101010101010101010101010101010101010101010101010a000000000000000101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010a0000000000000001010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101")
)

var (
	targetNumber = sc.U64(10)

	accountId, _ = types.NewAccountId(constants.OneAddress.FixedSequence...)

	targetHashAndNumber = TargetHashAndNumber{
		TargetHash: primitives.H256{
			FixedSequence: sc.FixedSequence[sc.U8]{
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1,
			},
		},
		TargetNumber: targetNumber,
	}

	signatureEd25519 = types.NewSignatureEd25519(sc.NewFixedSequence[sc.U8](
		64,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1,
	)...)

	tuple2HashSignature = Tuple2HashSignature{
		TargetHashAndNumber: targetHashAndNumber,
		SignatureEd25519:    signatureEd25519,
	}

	pubKey = primitives.Ed25519PublicKey(accountId)

	equivocation = Equivocation{
		RoundNumber: targetNumber,
		Identity:    pubKey,
		First:       tuple2HashSignature,
		Second:      tuple2HashSignature,
	}
)

func Test_TargetHashAndNumber_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetHashAndNumber.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, targetHashAndNumberBytes, buffer.Bytes())
}

func Test_TargetHashAndNumber_Bytes(t *testing.T) {
	assert.Equal(t, targetHashAndNumberBytes, targetHashAndNumber.Bytes())
}

func Test_DecodeTargetHashAndNumber(t *testing.T) {
	buffer := bytes.NewBuffer(targetHashAndNumberBytes)

	result, err := DecodeTargetHashAndNumber(buffer)

	assert.NoError(t, err)
	assert.Equal(t, targetHashAndNumber, result)
}

func Test_DecodeTargetHashAndNumber_Fails(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x0})

	_, err := DecodeTargetHashAndNumber(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_DecodeTargetHashAndNumber_Fails_To_Decode_Target_Number(t *testing.T) {
	buffer := bytes.NewBuffer(targetHashAndNumberBytes[:len(targetHashAndNumberBytes)-1])

	_, err := DecodeTargetHashAndNumber(buffer)

	assert.Error(t, err)
}

func Test_Tuple2HashSignature_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	tuple2HashSignature.Encode(buffer)

	assert.Equal(t, hashSignatureBytes, buffer.Bytes())
}

func Test_Tuple2HashSignature_Bytes(t *testing.T) {
	assert.Equal(t, hashSignatureBytes, tuple2HashSignature.Bytes())
}

func Test_DecodeTuple2HashSignature(t *testing.T) {
	buffer := bytes.NewBuffer(append(targetHashAndNumberBytes, signatureEd25519Bytes...))

	result, err := DecodeTuple2HashSignature(buffer)

	assert.NoError(t, err)
	assert.Equal(t, tuple2HashSignature, result)
}

func Test_DecodeTuple2HashSignature_Fails_To_Decode_HashAndNumber(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x0})

	_, err := DecodeTuple2HashSignature(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_DecodeTuple2HashSignature_Fails_To_Decode_Signature(t *testing.T) {
	buffer := bytes.NewBuffer(append(targetHashAndNumberBytes, 0x0))

	_, err := DecodeTuple2HashSignature(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_Equivocation_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := equivocation.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, equivocationBytes, buffer.Bytes())
}

func Test_Equivocation_Bytes(t *testing.T) {
	assert.Equal(t, equivocationBytes, equivocation.Bytes())
}

func Test_DecodeEquivocation(t *testing.T) {
	buffer := bytes.NewBuffer(equivocationBytes)

	result, err := DecodeEquivocation(buffer)

	assert.NoError(t, err)
	assert.Equal(t, equivocation, result)
}

func Test_DecodeEquivocation_Fails_To_Decode_Round(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x0})

	_, err := DecodeEquivocation(buffer)

	assert.Error(t, err)
}

func Test_DecodeEquivocation_Fails_To_Decode_Identity(t *testing.T) {
	buffer := bytes.NewBuffer(append(targetNumber.Bytes(), pubKey.Bytes()...))

	_, err := DecodeEquivocation(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_DecodeEquivocation_Fails_To_Decode_First(t *testing.T) {
	buffer := bytes.NewBuffer(append(append(targetNumber.Bytes(), pubKey.Bytes()...), 0x0))

	_, err := DecodeEquivocation(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_DecodeEquivocation_Fails_To_Decode_Second(t *testing.T) {
	buffer := bytes.NewBuffer(equivocationBytes[:len(equivocationBytes)-1])

	_, err := DecodeEquivocation(buffer)

	assert.Equal(t, io.EOF, err)
}
