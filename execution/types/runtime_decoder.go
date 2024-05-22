package types

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/LimeChain/gosemble/primitives/io"

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

type runtimeDecoder struct {
	modules           []types.Module
	extra             primitives.SignedExtra
	storage           io.Storage
	transactionBroker io.TransactionBroker
	logger            log.WarnLogger
}

func NewRuntimeDecoder(modules []types.Module, extra primitives.SignedExtra, storage io.Storage, transactionBroker io.TransactionBroker, logger log.WarnLogger) RuntimeDecoder {
	return runtimeDecoder{
		modules:           modules,
		extra:             extra,
		storage:           storage,
		transactionBroker: transactionBroker,
		logger:            logger,
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

	return NewUncheckedExtrinsic(sc.U8(version), extSignature, function, extra, rd.storage, rd.transactionBroker, rd.logger), nil
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

	function, err = function.DecodeArgs(buffer)
	if err != nil {
		return nil, err
	}

	return function, nil
}
