package main

import (
	"bytes"
	"testing"

	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/stretchr/testify/assert"
)

// TODO: in the test case "Commit_Then_Rollback" the host
// panics with "fatal error: exitsyscall: syscall frame is no longer valid"

func Test_Storage_Layer_Rollback_Then_Commit(t *testing.T) {
	rt, storage := newTestRuntime(t)

	metadata := runtimeMetadata(t, rt)

	call, err := ctypes.NewCall(metadata, "Testable.test", []byte{})
	assert.NoError(t, err)

	extrinsic := ctypes.NewExtrinsic(call)

	buffer := &bytes.Buffer{}
	encoder := cscale.NewEncoder(buffer)
	err = extrinsic.Encode(*encoder)
	assert.NoError(t, err)

	_, err = rt.Exec("BlockBuilder_apply_extrinsic", buffer.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, []byte{1}, (*storage).Get([]byte("testvalue")))
}
