package main

import (
	"bytes"
	"testing"
	"time"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_BlockBuilder_Inherent_Extrinsics(t *testing.T) {
	idata := gossamertypes.NewInherentData()
	time := time.Now().UnixMilli()
	err := idata.SetInherent(gossamertypes.Timstap0, uint64(time))
	assert.NoError(t, err)

	decoder := types.NewRuntimeDecoder(modules, newSignedExtra(modules), SudoIndex, ioStorage, ioTransactionBroker, log.NewLogger())

	rt, _ := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	expectedExtrinsicBytes := testhelpers.TimestampExtrinsicBytes(t, metadata, uint64(time))

	ienc, err := idata.Encode()
	assert.NoError(t, err)

	inherentExt, err := rt.Exec("BlockBuilder_inherent_extrinsics", ienc)
	assert.NoError(t, err)

	assert.NotNil(t, inherentExt)

	buffer := &bytes.Buffer{}
	buffer.Write([]byte{inherentExt[0]})

	totalInherents, err := sc.DecodeCompact[sc.U128](buffer)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), totalInherents.ToBigInt().Int64())
	buffer.Reset()

	buffer.Write(inherentExt[1:])
	extrinsic, err := decoder.DecodeUncheckedExtrinsic(buffer)
	assert.Nil(t, err)

	assert.Equal(t, expectedExtrinsicBytes, extrinsic.Bytes())
}
