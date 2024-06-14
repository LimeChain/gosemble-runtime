package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesAbridgedHostConfiguration, _ = hex.DecodeString("0100000002000000030000000400000005000000060000000700000008000000090000000a0000000b000000")
)

var (
	targetAbridgedHostConfiguration = AbridgedHostConfiguration{
		MaxCodeSize:                     1,
		MaxHeadDataSize:                 2,
		MaxUpwardQueueCount:             3,
		MaxUpwardQueueSize:              4,
		MaxUpwardMessageSize:            5,
		MaxUpwardMessageNumPerCandidate: 6,
		MaxHrmpMessageNumPerCandidate:   7,
		ValidationUpgradeCooldown:       8,
		ValidationUpgradeDelay:          9,
		AsyncBackingParams: AsyncBackingParams{
			MaxCandidateDepth:  10,
			AllowedAncestryLen: 11,
		},
	}
)

func Test_AbridgedHostConfiguration_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetAbridgedHostConfiguration.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, expectedBytesAbridgedHostConfiguration, buffer.Bytes())
}

func Test_AbridgedHostConfiguration_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesAbridgedHostConfiguration)

	result, err := DecodeAbridgeHostConfiguration(buf)
	assert.NoError(t, err)

	assert.Equal(t, targetAbridgedHostConfiguration, result)
}

func Test_AbridgedHostConfiguration_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesAbridgedHostConfiguration, targetAbridgedHostConfiguration.Bytes())
}
