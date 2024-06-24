package parachain

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func Test_NewDigestRelayParentStorageRoot(t *testing.T) {
	storageRoot := primitives.H256{FixedSequence: constants.ZeroAccountId.FixedSequence}
	number := sc.NewU32(5)

	expect := primitives.NewDigestItemConsensusMessage(
		sc.BytesToFixedSequenceU8(ConsensusTypeId[:]),
		sc.BytesToSequenceU8(sc.NewVaryingData(storageRoot, number).Bytes()),
	)

	result := NewDigestRelayParentStorageRoot(storageRoot, number)
	assert.Equal(t, expect, result)
}
