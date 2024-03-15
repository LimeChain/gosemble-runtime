package main

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/stretchr/testify/assert"
)

func Test_Metadata_Encoding_Success(t *testing.T) {
	runtime, _ := newTestRuntime(t)
	gossamerMetadata := runtimeMetadata(t, runtime)

	bMetadata, err := runtime.Metadata()
	assert.NoError(t, err)

	buffer := bytes.NewBuffer(bMetadata)

	// Decode Compact Length
	_, err = sc.DecodeCompact[sc.U128](buffer)
	assert.Nil(t, err)

	// Copy bytes for assertion after re-encode.
	bMetadataCopy := make([]byte, buffer.Len())
	copy(bMetadataCopy, buffer.Bytes())

	metadata, err := types.DecodeMetadata(buffer)
	assert.Nil(t, err)

	// Assert encoding of previously decoded
	assert.Equal(t, bMetadataCopy, metadata.Bytes())

	// Encode gossamer Metadata
	bGossamerMetadata, err := codec.Encode(gossamerMetadata)
	assert.NoError(t, err)

	assert.Equal(t, metadata.Bytes(), bGossamerMetadata)
}

func Test_Metadata_Versions_Correct_Versions(t *testing.T) {
	runtime, _ := newTestRuntime(t)

	metadataVersions, err := runtime.Exec("Metadata_metadata_versions", []byte{})
	assert.NoError(t, err)

	buffer := &bytes.Buffer{}
	buffer.Write(metadataVersions)

	versions, err := sc.DecodeSequence[sc.U32](buffer)
	assert.Nil(t, err)
	buffer.Reset()

	expectedVersions := sc.Sequence[sc.U32]{
		sc.U32(types.MetadataVersion14),
		sc.U32(types.MetadataVersion15),
	}

	assert.Equal(t, versions, expectedVersions)
}

func Test_Metadata_At_Version_14(t *testing.T) {
	runtime, _ := newTestRuntime(t)
	gossamerMetadata := runtimeMetadata(t, runtime)

	version14 := sc.U32(types.MetadataVersion14)

	bMetadata, err := runtime.Exec("Metadata_metadata_at_version", version14.Bytes())
	assert.NoError(t, err)

	resultOptionMetadataBuffer := bytes.NewBuffer(bMetadata)

	optionMetadata, err := sc.DecodeOptionWith[sc.Sequence[sc.U8]](resultOptionMetadataBuffer, sc.DecodeSequence[sc.U8])
	assert.Nil(t, err)

	metadataV14Bytes := optionMetadata.Value.Bytes()

	buffer := bytes.NewBuffer(metadataV14Bytes)

	// Decode Compact Length
	_, err = sc.DecodeCompact[sc.U128](buffer)
	assert.Nil(t, err)

	// Copy bytes for assertion after re-encode.
	bMetadataCopy := make([]byte, buffer.Len())
	copy(bMetadataCopy, buffer.Bytes())

	metadata, err := types.DecodeMetadata(buffer)
	assert.Nil(t, err)

	assert.Equal(t, bMetadataCopy, metadata.Bytes())

	bGossamerMetadata, err := codec.Encode(gossamerMetadata)
	assert.NoError(t, err)

	assert.Equal(t, metadata.Bytes(), bGossamerMetadata)
}

func Test_Metadata_At_Version_15(t *testing.T) {
	runtime, _ := newTestRuntime(t)

	version15 := sc.U32(types.MetadataVersion15)

	bMetadata, err := runtime.Exec("Metadata_metadata_at_version", version15.Bytes())
	assert.NoError(t, err)

	resultOptionMetadataBuffer := bytes.NewBuffer(bMetadata)

	optionMetadata, err := sc.DecodeOptionWith[sc.Sequence[sc.U8]](resultOptionMetadataBuffer, sc.DecodeSequence[sc.U8])
	assert.Nil(t, err)

	metadataV15Bytes := optionMetadata.Value.Bytes()

	buffer := bytes.NewBuffer(metadataV15Bytes)

	// Decode Compact Length
	_, err = sc.DecodeCompact[sc.U128](buffer)
	assert.Nil(t, err)

	bMetadataCopy := make([]byte, buffer.Len())
	copy(bMetadataCopy, buffer.Bytes())

	metadata, err := types.DecodeMetadata(buffer)
	assert.Nil(t, err)

	assert.Equal(t, bMetadataCopy, metadata.Bytes())
}

func Test_Metadata_At_Version_UnsupportedVersion(t *testing.T) {
	runtime, _ := newTestRuntime(t)

	unsupportedVersion := sc.U32(10)

	bMetadata, err := runtime.Exec("Metadata_metadata_at_version", unsupportedVersion.Bytes())
	assert.NoError(t, err)

	buffer := bytes.NewBuffer(bMetadata)

	result, err := sc.DecodeOption[types.Metadata](buffer)
	assert.Nil(t, err)

	expectedResult := sc.Option[types.Metadata]{
		HasValue: sc.Bool(false),
	}

	assert.Equal(t, result.Bytes(), expectedResult.Bytes())
}
