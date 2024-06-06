package session

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	membershipProof = MembershipProof{
		SessionIndex: 1,
		TrieNodes: sc.Sequence[sc.Sequence[sc.U8]]{
			sc.Sequence[sc.U8]{2, 3},
			sc.Sequence[sc.U8]{4, 5, 6},
		},
		Validators: 7,
	}

	membershipProofBytes, _ = hex.DecodeString("01000000080802030c04050607000000")
)

func Test_MembershipProof_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := membershipProof.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, membershipProofBytes, buffer.Bytes())
}

func Test_MembershipProof_Bytes(t *testing.T) {
	assert.Equal(t, membershipProofBytes, membershipProof.Bytes())
}

func Test_DecodeMembershipProof(t *testing.T) {
	buffer := bytes.NewBuffer(membershipProofBytes)

	result, err := DecodeMembershipProof(buffer)

	assert.NoError(t, err)
	assert.Equal(t, membershipProof, result)
}

func Test_DecodeMembershipProof_Fails_To_Decode_SessionIndex(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x0})

	_, err := DecodeMembershipProof(buffer)

	assert.Error(t, err)
}

func Test_DecodeMembershipProof_Fails_To_Decode_TrieNodes(t *testing.T) {
	buffer := bytes.NewBuffer(membershipProofBytes[:4])

	_, err := DecodeMembershipProof(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_DecodeMembershipProof_Fails_To_Decode_ValidatorCount(t *testing.T) {
	buffer := bytes.NewBuffer(membershipProofBytes[:len(membershipProofBytes)-1])

	_, err := DecodeMembershipProof(buffer)

	assert.Error(t, err)
}

func Test_MembershipProof_Session(t *testing.T) {
	assert.Equal(t, sc.U32(1), membershipProof.Session())
}

func Test_MembershipProof_ValidatorCount(t *testing.T) {
	assert.Equal(t, sc.U32(7), membershipProof.ValidatorCount())
}
