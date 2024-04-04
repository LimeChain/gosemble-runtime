package session

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GenesisConfig_CreateDefaultConfig(t *testing.T) {
	target := setupModule()

	expectedGc := []byte("{\"session\":{\"keys\":[]}}")

	gc, err := target.CreateDefaultConfig()
	assert.NoError(t, err)
	assert.Equal(t, expectedGc, gc)
}
