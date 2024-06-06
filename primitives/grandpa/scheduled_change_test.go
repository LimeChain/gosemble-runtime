package grandpa

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	authorities = sc.Sequence[types.Authority]{
		{
			Id:     constants.ZeroAccountId,
			Weight: sc.U64(1),
		},
		{
			Id:     constants.OneAccountId,
			Weight: sc.U64(2),
		},
	}

	scheduledChange = ScheduledChange{
		NextAuthorities: authorities,
		Delay:           sc.U64(123456),
	}
)

var (
	scheduledChangeBytes, err = hex.DecodeString("08000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000001020000000000000040e2010000000000")
)

func Test_ScheduledChange_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	scheduledChange.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, scheduledChangeBytes, buffer.Bytes())
}

func Test_ScheduledChange_Bytes(t *testing.T) {
	assert.NoError(t, err)
	assert.Equal(t, scheduledChangeBytes, scheduledChange.Bytes())
}

func Test_DecodeScheduledChange(t *testing.T) {
	buffer := bytes.NewBuffer(scheduledChangeBytes)

	result, err := DecodeScheduledChange(buffer)

	assert.NoError(t, err)
	assert.Equal(t, scheduledChange, result)
}

func Test_DecodeScheduledChange_Failing_To_Decode_Authorities(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x08})

	_, err := DecodeScheduledChange(buffer)

	assert.Equal(t, io.EOF, err)
}

func Test_DecodeScheduledChange_Failing_To_Decode_Delay(t *testing.T) {
	buffer := bytes.NewBuffer(scheduledChangeBytes[:len(scheduledChangeBytes)-1])

	_, err := DecodeScheduledChange(buffer)

	assert.Error(t, err)
}
