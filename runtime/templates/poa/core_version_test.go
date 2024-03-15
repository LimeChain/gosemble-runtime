package main

import (
	"bytes"
	"testing"

	runtimetypes "github.com/ChainSafe/gossamer/lib/runtime"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/stretchr/testify/assert"
)

func Test_CoreVersion(t *testing.T) {
	rt, _ := newTestRuntime(t)

	versionEncBytes, err := rt.Exec("Core_version", []byte{})
	assert.NoError(t, err)

	buffer := bytes.NewBuffer(versionEncBytes)

	dec := scale.NewDecoder(buffer)
	runtimeVersion := runtimetypes.Version{}
	err = dec.Decode(&runtimeVersion)
	assert.NoError(t, err)
	assert.Equal(t, "node-template", string(runtimeVersion.SpecName))
	assert.Equal(t, "node-template", string(runtimeVersion.ImplName))
	assert.Equal(t, uint32(1), runtimeVersion.AuthoringVersion)
	assert.Equal(t, uint32(100), runtimeVersion.SpecVersion)
	assert.Equal(t, uint32(1), runtimeVersion.ImplVersion)
	assert.Equal(t, uint32(1), runtimeVersion.TransactionVersion)
	assert.Equal(t, uint8(1), runtimeVersion.StateVersion)
}
