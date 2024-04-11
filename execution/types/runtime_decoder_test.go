package types

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	moduleOneIdx           = sc.U8(0)
	functionIdx            = sc.U8(0)
	signedExtrinsicVersion = 132
	defaultSudoIndex       = 0
	sudoIndex              = sc.U8(7)
)

var (
	signedExtrinsicBytes = []byte{
		byte(signedExtrinsicVersion),                                                                      // version
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // signer
		0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, // signature,
		// extra
		uint8(moduleOneIdx), uint8(functionIdx), // call
	}

	header = primitives.Header{
		ParentHash: primitives.Blake2bHash{
			FixedSequence: sc.BytesToFixedSequenceU8(parentHash)},
		Number:         5,
		StateRoot:      primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(stateRoot)},
		ExtrinsicsRoot: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(extrinsicsRoot)},
		Digest:         primitives.NewDigest(sc.Sequence[primitives.DigestItem]{}),
	}

	parentHash      = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	stateRoot       = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355b").ToBytes()
	extrinsicsRoot  = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355a").ToBytes()
	moduleFunctions = map[sc.U8]primitives.Call{}
)

var (
	mockModuleOne  *mocks.Module
	mockCallOne    *mocks.Call
	mockDecodeCall = mock.AnythingOfType("func(*bytes.Buffer) (types.Call, error)")
	logger         = log.NewLogger()
)

func Test_RuntimeDecoder_New(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)
	expect := runtimeDecoder{
		modules:   []primitives.Module{mockModuleOne},
		extra:     mockSignedExtra,
		sudoIndex: sc.U8(0),
		logger:    logger,
	}

	assert.Equal(t, expect, target)
}

func Test_RuntimeDecoder_DecodeBlock_ZeroExtrinsicsEmptyBody(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)

	lenExtrinsics := sc.ToCompact(0).Bytes()
	buff := bytes.NewBuffer(append(header.Bytes(), lenExtrinsics...))

	resultBlock, err := target.DecodeBlock(buff)
	assert.NoError(t, err)

	expectedBlock := NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{})

	assert.Equal(t, expectedBlock, resultBlock)
}

func Test_RuntimeDecoder_DecodeBlock_Single_Extrinsic(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)

	moduleFunctions[0] = mockCallOne

	buffExtrinsic := bytes.NewBuffer(append(sc.ToCompact(len(signedExtrinsicBytes)).Bytes(), signedExtrinsicBytes...))

	lenExtrinsics := sc.ToCompact(1).Bytes()
	decodeBlockBytes := append(lenExtrinsics, buffExtrinsic.Bytes()...)
	decodeBlockBytes = append(header.Bytes(), decodeBlockBytes...)

	decodeBlockBuff := bytes.NewBuffer(decodeBlockBytes)

	mockSignedExtra.On("DeepCopy").Return(mockSignedExtra)
	mockModuleOne.On("GetIndex").Return(moduleOneIdx)
	mockSignedExtra.On("Decode", mock.Anything).Return()
	mockModuleOne.On("Functions").Return(moduleFunctions)
	mockCallOne.On("DecodeArgs", decodeBlockBuff).Return(mockCallOne, nil)
	result, err := target.DecodeBlock(decodeBlockBuff)
	assert.NoError(t, err)

	extrinsics := sc.Sequence[primitives.UncheckedExtrinsic]{
		NewUncheckedExtrinsic(sc.U8(signedExtrinsicVersion), extrinsicSignature, mockCallOne, mockSignedExtra, logger),
	}

	expectedBlock := NewBlock(header, extrinsics)

	assert.Equal(t, expectedBlock.Header().Number, result.Header().Number)
	assert.Equal(t, expectedBlock.Header().Digest, result.Header().Digest)
	assert.Equal(t, expectedBlock.Header().ParentHash, result.Header().ParentHash)
	assert.Equal(t, expectedBlock.Header().ExtrinsicsRoot, result.Header().ExtrinsicsRoot)
	assert.Equal(t, expectedBlock.Header().StateRoot, result.Header().StateRoot)

	assert.Equal(t, expectedBlock.Extrinsics()[0].Signature(), result.Extrinsics()[0].Signature())
	assert.Equal(t, expectedBlock.Extrinsics()[0].Function(), result.Extrinsics()[0].Function())
	assert.Equal(t, expectedBlock.Extrinsics()[0].Extra(), result.Extrinsics()[0].Extra())
	assert.Equal(t, expectedBlock.Extrinsics()[0].IsSigned(), result.Extrinsics()[0].IsSigned())

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
	mockCallOne.AssertCalled(t, "DecodeArgs", decodeBlockBuff)
	mockSignedExtra.AssertCalled(t, "Decode", mock.Anything)
}

func Test_RuntimeDecoder_DecodeBlock_Multiple_Extrinsics(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)
	moduleFunctions[0] = mockCallOne

	totalExtrinsicsInBlock := 10
	lenSignedExtrinsicBytes := sc.ToCompact(len(signedExtrinsicBytes)).Bytes()

	signedExtrinsicBytes := append(lenSignedExtrinsicBytes, signedExtrinsicBytes...)
	allExtrinsicsBytes := signedExtrinsicBytes
	assert.Equal(t, allExtrinsicsBytes, signedExtrinsicBytes)
	for i := 0; i < totalExtrinsicsInBlock-1; i++ {
		allExtrinsicsBytes = append(signedExtrinsicBytes, allExtrinsicsBytes...)
	}

	lenExtrinsics := sc.ToCompact(totalExtrinsicsInBlock).Bytes()
	decodeBlockBytes := append(lenExtrinsics, allExtrinsicsBytes...)
	decodeBlockBytes = append(header.Bytes(), decodeBlockBytes...)

	buff := bytes.NewBuffer(decodeBlockBytes)

	mockSignedExtra.On("DeepCopy").Return(mockSignedExtra)
	mockModuleOne.On("GetIndex").Return(moduleOneIdx)
	mockSignedExtra.On("Decode", mock.Anything).Return()
	mockModuleOne.On("Functions").Return(moduleFunctions)
	mockCallOne.On("DecodeArgs", buff).Return(mockCallOne, nil)
	result, err := target.DecodeBlock(buff)
	assert.NoError(t, err)

	extrinsics := sc.Sequence[primitives.UncheckedExtrinsic]{}
	for i := 0; i < totalExtrinsicsInBlock; i++ {
		extrinsics = append(extrinsics, NewUncheckedExtrinsic(sc.U8(signedExtrinsicVersion), extrinsicSignature, mockCallOne, mockSignedExtra, logger))
	}

	expectedBlock := NewBlock(header, extrinsics)

	assert.Equal(t, expectedBlock.Header().Number, result.Header().Number)
	assert.Equal(t, expectedBlock.Header().Digest, result.Header().Digest)
	assert.Equal(t, expectedBlock.Header().ParentHash, result.Header().ParentHash)
	assert.Equal(t, expectedBlock.Header().ExtrinsicsRoot, result.Header().ExtrinsicsRoot)
	assert.Equal(t, expectedBlock.Header().StateRoot, result.Header().StateRoot)

	for i := 0; i < totalExtrinsicsInBlock; i++ {
		assert.Equal(t, expectedBlock.Extrinsics()[i].Signature(), result.Extrinsics()[i].Signature())
		assert.Equal(t, expectedBlock.Extrinsics()[i].Function(), result.Extrinsics()[i].Function())
		assert.Equal(t, expectedBlock.Extrinsics()[i].Extra(), result.Extrinsics()[i].Extra())
		assert.Equal(t, expectedBlock.Extrinsics()[i].IsSigned(), result.Extrinsics()[i].IsSigned())
	}

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
	mockCallOne.AssertCalled(t, "DecodeArgs", buff)
	mockSignedExtra.AssertCalled(t, "Decode", mock.Anything)
}

func Test_RuntimeDecoder_DecodeUncheckedExtrinsic_Unsigned(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)
	moduleFunctions[0] = mockCallOne

	unsignedExtrBytes := append(moduleOneIdx.Bytes(), functionIdx.Bytes()...)
	unsignedExtrBytes = append(sc.U8(ExtrinsicFormatVersion).Bytes(), unsignedExtrBytes...)
	unsignedExtrBytes = append(sc.ToCompact(len(unsignedExtrBytes)).Bytes(), unsignedExtrBytes...)

	buff := bytes.NewBuffer(unsignedExtrBytes)

	mockSignedExtra.On("DeepCopy").Return(mockSignedExtra)
	mockModuleOne.On("GetIndex").Return(moduleOneIdx)
	mockModuleOne.On("Functions").Return(moduleFunctions)
	mockCallOne.On("DecodeArgs", buff).Return(mockCallOne, nil)

	result, err := target.DecodeUncheckedExtrinsic(buff)
	assert.NoError(t, err)

	expectedUnsignedExtrinsic := NewUncheckedExtrinsic(version, sc.Option[primitives.ExtrinsicSignature]{}, mockCallOne, mockSignedExtra, logger)

	assert.Equal(t, expectedUnsignedExtrinsic.IsSigned(), result.IsSigned())

	assert.Equal(t, expectedUnsignedExtrinsic.Signature(), result.Signature())
	assert.Equal(t, expectedUnsignedExtrinsic.Function(), result.Function())
	assert.Equal(t, expectedUnsignedExtrinsic.Extra(), result.Extra())

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
	mockCallOne.AssertCalled(t, "DecodeArgs", buff)
}

func Test_RuntimeDecoder_DecodeUncheckedExtrinsic_Signed(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)

	mockSignedExtra.On("Decode", mock.Anything).Return()

	buff := bytes.NewBuffer(append(sc.ToCompact(len(signedExtrinsicBytes)).Bytes(), signedExtrinsicBytes...))

	moduleFunctions[0] = mockCallOne

	mockSignedExtra.On("DeepCopy").Return(mockSignedExtra)
	mockModuleOne.On("GetIndex").Return(moduleOneIdx)
	mockModuleOne.On("Functions").Return(moduleFunctions)
	mockCallOne.On("DecodeArgs", buff).Return(mockCallOne, nil)

	result, err := target.DecodeUncheckedExtrinsic(buff)
	assert.NoError(t, err)

	expectedSignedExtrinsicsBytesAfterDecode := NewUncheckedExtrinsic(sc.U8(signedExtrinsicVersion), extrinsicSignature, mockCallOne, mockSignedExtra, logger)

	assert.Equal(t, expectedSignedExtrinsicsBytesAfterDecode.IsSigned(), result.IsSigned())

	assert.Equal(t, expectedSignedExtrinsicsBytesAfterDecode.Signature(), result.Signature())
	assert.Equal(t, expectedSignedExtrinsicsBytesAfterDecode.Function(), result.Function())
	assert.Equal(t, expectedSignedExtrinsicsBytesAfterDecode.Extra(), result.Extra())

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
	mockCallOne.AssertCalled(t, "DecodeArgs", buff)
}

func Test_RuntimeDecoder_DecodeUncheckedExtrinsic_InvalidExtrinsicVersion(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)

	invalidExtrinsicVersion := byte(99)

	signedExtrinsicBytesInvalid := make([]byte, len(signedExtrinsicBytes))
	copy(signedExtrinsicBytesInvalid, signedExtrinsicBytes)

	signedExtrinsicBytesInvalid[0] = invalidExtrinsicVersion

	buff := bytes.NewBuffer(append(sc.ToCompact(len(signedExtrinsicBytesInvalid)).Bytes(), signedExtrinsicBytesInvalid...))

	_, err := target.DecodeUncheckedExtrinsic(buff)
	assert.Equal(t, errInvalidExtrinsicVersion, err)
}

func Test_RuntimeDecoder_DecodeUncheckedExtrinsic_InvalidLengthPrefix(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)

	mockSignedExtra.On("Decode", mock.Anything).Return()

	invalidExpectedLength := sc.ToCompact(len(signedExtrinsicBytes) - 1)

	buff := bytes.NewBuffer(append(invalidExpectedLength.Bytes(), signedExtrinsicBytes...))

	moduleFunctions[0] = mockCallOne

	mockSignedExtra.On("DeepCopy").Return(mockSignedExtra)
	mockModuleOne.On("GetIndex").Return(moduleOneIdx)
	mockModuleOne.On("Functions").Return(moduleFunctions)
	mockCallOne.On("DecodeArgs", buff).Return(mockCallOne, nil)

	_, err := target.DecodeUncheckedExtrinsic(buff)
	assert.Equal(t, errInvalidLengthPrefix, err)

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
	mockCallOne.AssertCalled(t, "DecodeArgs", buff)
}

func Test_RuntimeDecoder_DecodeCall_Module_Not_Exists(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)

	idxModuleNotExists := sc.U8(10)

	callBytes := []byte{
		uint8(idxModuleNotExists), uint8(functionIdx),
	}

	buf := bytes.NewBuffer(callBytes)
	mockModuleOne.On("GetIndex").Return(moduleOneIdx)

	mod, err := target.DecodeCall(buf)

	assert.Equal(t, err.Error(), "module with index ["+strconv.Itoa(int(idxModuleNotExists))+"] not found.")
	assert.Nil(t, mod)

	mockModuleOne.AssertCalled(t, "GetIndex")
}

func Test_RuntimeDecoder_DecodeCall_Function_Not_Exists(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)

	idxFunctionNotExists := sc.U8(10)

	callBytes := []byte{
		uint8(moduleOneIdx), uint8(idxFunctionNotExists),
	}

	buf := bytes.NewBuffer(callBytes)
	moduleFunctions[0] = mockCallOne

	mockModuleOne.On("GetIndex").Return(moduleOneIdx)
	mockModuleOne.On("Functions").Return(moduleFunctions)

	_, err := target.DecodeCall(buf)
	assert.Equal(t, fmt.Errorf("function index [%d] for module [%d] not found", idxFunctionNotExists, moduleOneIdx), err)

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
}

func Test_RuntimeDecoder_DecodeCall(t *testing.T) {
	target := setupRuntimeDecoder(defaultSudoIndex)

	args := sc.NewVaryingData(sc.U8(1), sc.U8(2), sc.U8(3))

	callBytes := []byte{
		uint8(moduleOneIdx), uint8(functionIdx),
	}

	callBytes = append(callBytes, args.Bytes()...)

	buf := bytes.NewBuffer(callBytes)
	moduleFunctions[0] = mockCallOne

	mockModuleOne.On("GetIndex").Return(moduleOneIdx)
	mockModuleOne.On("Functions").Return(moduleFunctions)
	mockCallOne.On("DecodeArgs", buf).Run(func(args mock.Arguments) {
		buf := args.Get(0).(*bytes.Buffer)
		// reading 3 bytes for the 3 arguments
		buf.ReadByte()
		buf.ReadByte()
		buf.ReadByte()
	}).Return(mockCallOne, nil)

	res, err := target.DecodeCall(buf)
	assert.NoError(t, err)
	assert.Equal(t, mockCallOne, res)

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
	mockCallOne.AssertCalled(t, "DecodeArgs", buf)
}

func Test_RuntimeDecoder_DecodeSudoCall_InvalidType(t *testing.T) {
	target := setupRuntimeDecoder(sudoIndex)

	args := sc.NewVaryingData(sc.U8(1), sc.U8(2), sc.U8(3))

	callBytes := []byte{
		uint8(sudoIndex), uint8(functionIdx),
	}

	callBytes = append(callBytes, args.Bytes()...)

	buf := bytes.NewBuffer(callBytes)
	moduleFunctions[0] = mockCallOne

	mockModuleOne.On("GetIndex").Return(sudoIndex)
	mockModuleOne.On("Functions").Return(moduleFunctions)

	res, err := target.DecodeCall(buf)
	assert.Equal(t, errors.New("function index [0] for module [7] does not implement sudo decoder"), err)
	assert.Nil(t, res)

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
	mockCallOne.AssertNotCalled(t, "DecodeSudoArgs")
}

func Test_RuntimeDecoder_DecodeCall_Sudo(t *testing.T) {
	target := setupRuntimeDecoder(sudoIndex)
	mockCallSudo := new(mocks.CallSudo)

	args := sc.NewVaryingData(sudoIndex, sc.U8(2), sc.U8(3))

	callBytes := append(sudoIndex.Bytes(), functionIdx.Bytes()...)

	callBytes = append(callBytes, args.Bytes()...)

	buf := bytes.NewBuffer(callBytes)
	moduleFunctions[0] = mockCallSudo

	mockModuleOne.On("GetIndex").Return(sudoIndex)
	mockModuleOne.On("Functions").Return(moduleFunctions)
	mockCallSudo.On("DecodeSudoArgs", buf, mockDecodeCall).Run(func(args mock.Arguments) {
		buf := args.Get(0).(*bytes.Buffer)
		// reading 3 bytes for the 3 arguments
		buf.ReadByte()
		buf.ReadByte()
		buf.ReadByte()
	}).Return(mockCallSudo, nil)

	res, err := target.DecodeCall(buf)
	assert.NoError(t, err)
	assert.Equal(t, mockCallSudo, res)

	mockModuleOne.AssertCalled(t, "GetIndex")
	mockModuleOne.AssertCalled(t, "Functions")
	mockCallSudo.AssertCalled(t, "DecodeSudoArgs", buf, mockDecodeCall)
}

func setupRuntimeDecoder(sudoIndex sc.U8) RuntimeDecoder {
	mockModuleOne = new(mocks.Module)

	mockCallOne = new(mocks.Call)

	mockSignedExtra = new(mocks.SignedExtra)

	extrinsicSignature = sc.NewOption[primitives.ExtrinsicSignature](
		primitives.ExtrinsicSignature{
			Signer:    signer,
			Signature: signatureEd25519,
			Extra:     mockSignedExtra,
		},
	)

	apis := []primitives.Module{mockModuleOne}

	return NewRuntimeDecoder(apis, mockSignedExtra, sudoIndex, logger)
}
