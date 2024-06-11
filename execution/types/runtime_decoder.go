package types

import (
	"bytes"
	"errors"
	"fmt"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	errInvalidExtrinsicVersion = errors.New("invalid Extrinsic version")
	errInvalidLengthPrefix     = errors.New("invalid length prefix")
)

type RuntimeDecoder interface {
	DecodeBlock(buffer *bytes.Buffer) (primitives.Block, error)
	DecodeUncheckedExtrinsic(buffer *bytes.Buffer) (primitives.UncheckedExtrinsic, error)
	DecodeCall(buffer *bytes.Buffer) (primitives.Call, error)
}

type SudoDecoder interface {
	DecodeSudoArgs(buffer *bytes.Buffer, decodeCallFunc func(buffer *bytes.Buffer) (primitives.Call, error)) (primitives.Call, error)
}

type runtimeDecoder struct {
	modules   []types.Module
	extra     primitives.SignedExtra
	sudoIndex sc.U8 // Used for additional decoding related to Sudo calls. Default is 0 and not considered a valid sudo index.
	logger    log.RuntimeLogger
}

func NewRuntimeDecoder(modules []types.Module, extra primitives.SignedExtra, sudoIndex sc.U8, logger log.RuntimeLogger) RuntimeDecoder {
	return runtimeDecoder{
		modules:   modules,
		extra:     extra,
		sudoIndex: sudoIndex,
		logger:    logger,
	}
}

func (rd runtimeDecoder) DecodeBlock(buffer *bytes.Buffer) (primitives.Block, error) {
	header, err := primitives.DecodeHeader(buffer)
	if err != nil {
		return nil, err
	}
	compact, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return nil, err
	}
	length := compact.ToBigInt().Int64()
	extrinsics := make([]types.UncheckedExtrinsic, length)

	for i := 0; i < len(extrinsics); i++ {
		extrinsic, err := rd.DecodeUncheckedExtrinsic(buffer)
		if err != nil {
			return nil, err
		}
		extrinsics[i] = extrinsic
	}

	return NewBlock(header, extrinsics), nil
}

func (rd runtimeDecoder) DecodeUncheckedExtrinsic(buffer *bytes.Buffer) (primitives.UncheckedExtrinsic, error) {
	// This is a little more complicated than usual since the binary format must be compatible
	// with SCALE's generic `Vec<u8>` type. Basically this just means accepting that there
	// will be a prefix of vector length.
	expectedLenCompact, err := sc.DecodeCompact[sc.U64](buffer)
	if err != nil {
		return nil, err
	}
	expectedLength := expectedLenCompact.ToBigInt().Int64()
	beforeLength := buffer.Len()

	version, _ := buffer.ReadByte()

	if version&ExtrinsicUnmaskVersion != ExtrinsicFormatVersion {
		return nil, errInvalidExtrinsicVersion
	}

	extra := rd.extra.DeepCopy()

	var extSignature sc.Option[primitives.ExtrinsicSignature]
	isSigned := version&ExtrinsicBitSigned != 0
	if isSigned {
		sig, err := primitives.DecodeExtrinsicSignature(extra, buffer)
		if err != nil {
			return nil, err
		}
		extSignature = sc.NewOption[primitives.ExtrinsicSignature](sig)
	}

	// Decodes the dispatch call, including its arguments.
	function, err := rd.DecodeCall(buffer)
	if err != nil {
		return nil, err
	}

	afterLength := buffer.Len()

	if int(expectedLength) != beforeLength-afterLength {
		return nil, errInvalidLengthPrefix
	}

	return NewUncheckedExtrinsic(sc.U8(version), extSignature, function, extra, rd.logger), nil
}

func (rd runtimeDecoder) DecodeCall(buffer *bytes.Buffer) (primitives.Call, error) {
	moduleIndex, err := sc.DecodeU8(buffer)
	if err != nil {
		return nil, err
	}

	functionIndex, err := sc.DecodeU8(buffer)
	if err != nil {
		return nil, err
	}

	module, err := primitives.GetModule(moduleIndex, rd.modules)
	if err != nil {
		return nil, err
	}

	function, ok := module.Functions()[functionIndex]
	if !ok {
		return nil, fmt.Errorf("function index [%d] for module [%d] not found", functionIndex, moduleIndex)
	}

	if rd.isSudoCall(moduleIndex) {
		sudoCall, ok := function.(SudoDecoder)
		if !ok {
			return nil, fmt.Errorf("function index [%d] for module [%d] does not implement sudo decoder", functionIndex, moduleIndex)
		}
		function, err = sudoCall.DecodeSudoArgs(buffer, rd.DecodeCall)
	} else {
		function, err = function.DecodeArgs(buffer)
	}
	if err != nil {
		return nil, err
	}

	return function, nil
}

// Checks if the current module index is Sudo.
func (rd runtimeDecoder) isSudoCall(moduleIndex sc.U8) bool {
	if rd.sudoIndex != 0 && moduleIndex == rd.sudoIndex {
		return true
	}

	return false
}
