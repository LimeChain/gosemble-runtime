package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedDispatchInfoBytes, _ = hex.DecodeString("04080201")

	targetDispatchInfo = DispatchInfo{
		Weight:  WeightFromParts(1, 2),
		Class:   NewDispatchClassMandatory(),
		PaysFee: PaysNo,
	}
)

func Test_DispatchInfo_Encode(t *testing.T) {
	buf := &bytes.Buffer{}

	err := targetDispatchInfo.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedDispatchInfoBytes, buf.Bytes())
}

func Test_DispatchInfo_Bytes(t *testing.T) {
	assert.Equal(t, expectedDispatchInfoBytes, targetDispatchInfo.Bytes())
}

func Test_DecodeDispatchInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	buf.Write(expectedDispatchInfoBytes)

	result, err := DecodeDispatchInfo(buf)
	assert.NoError(t, err)

	assert.Equal(t, targetDispatchInfo, result)
}

func Test_GetDispatchInfo(t *testing.T) {
	call := testCall{}

	result := GetDispatchInfo(call)

	assert.Equal(t, DispatchInfo{
		Weight:  WeightFromParts(3, 4),
		Class:   NewDispatchClassNormal(),
		PaysFee: PaysYes,
	}, result)
}

func Test_IsMendatory(t *testing.T) {
	for _, tt := range []struct {
		name        string
		class       DispatchClass
		expectedErr error
		expectedRes bool
	}{
		{
			name:        "is mendatory",
			class:       NewDispatchClassMandatory(),
			expectedRes: true,
		},
		{
			name:  "not mendatory",
			class: NewDispatchClassNormal(),
		},
		{
			name:        "err invalid DispatchClass",
			class:       DispatchClass{},
			expectedErr: newTypeError("DispatchClass"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			res, err := DispatchInfo{Class: tt.class}.IsMandatory()
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, sc.Bool(tt.expectedRes), res)
		})
	}
}
