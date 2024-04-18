package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/parachain_system"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
	"reflect"
)

const (
	ApiModuleName = "Parachain"
	apiVersion    = 1
)

// Module implements the Parachain `validate_block` Runtime API function
type Module struct {
	parachainSystem parachain_system.Module
	runtimeDecoder  types.RuntimeDecoder
	hashing         io.Hashing
	logger          log.Logger
	memUtils        utils.WasmMemoryTranslator
}

func New(parachainSystem parachain_system.Module, runtimeDecoder types.RuntimeDecoder, logger log.Logger) Module {
	return Module{
		parachainSystem: parachainSystem,
		runtimeDecoder:  runtimeDecoder,
		hashing:         io.NewHashing(),
		logger:          logger,
		memUtils:        utils.NewMemoryTranslator(),
	}
}

// Name returns the name of the api module
func (m Module) Name() string {
	return ApiModuleName
}

// Item returns the first 8 bytes of the Blake2b hash of the name and version of the api module.
func (m Module) Item() primitives.ApiItem {
	hash := hashing.MustBlake2b8([]byte(ApiModuleName))
	return primitives.NewApiItem(hash, apiVersion)
}

func (m Module) ValidateBlock(dataPtr int32, dataLen int32) int64 {
	b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	buffer := bytes.NewBuffer(b)

	validationData, err := parachain.DecodeValidationParams(buffer)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	blockData, err := decodeParachainBlockData(m.runtimeDecoder, validationData.BlockData)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	parentHeader, err := primitives.DecodeHeader(bytes.NewBuffer(sc.SequenceU8ToBytes(validationData.ParentHead)))
	if err != nil {
		m.logger.Critical(err.Error())
	}

	parentHeaderHash := sc.BytesToFixedSequenceU8(m.hashing.Blake256(parentHeader.Bytes()))
	if reflect.DeepEqual(parentHeaderHash, blockData.Block.Header().ParentHash) {
		m.logger.Critical("invalid parent hash")
	}

	// TODO: execute block

	collationInfo, err := m.parachainSystem.CollectCollationInfo(blockData.Block.Header())
	if err != nil {
		m.logger.Critical(err.Error())
	}

	// TODO: Fields are the same as collation info, but in different encoding order
	result := parachain.ValidationResult{
		HeadData:                  collationInfo.HeadData,
		NewValidationCode:         collationInfo.ValidationCode,
		UpwardMessages:            collationInfo.UpwardMessages,
		HorizontalMessages:        collationInfo.HorizontalMessages,
		ProcessedDownwardMessages: collationInfo.ProcessedDownwardMessages,
		HrmpWatermark:             collationInfo.HrmpWatermark,
	}

	return m.memUtils.BytesToOffsetAndSize(result.Bytes())
}

func decodeParachainBlockData(runtimeDecoder types.RuntimeDecoder, blockData sc.Sequence[sc.U8]) (parachain.BlockData, error) {
	buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(blockData))

	block, err := runtimeDecoder.DecodeBlock(buffer)
	if err != nil {
		return parachain.BlockData{}, err
	}

	compactProofs, err := sc.DecodeSequenceWith(buffer, sc.DecodeSequence[sc.U8])
	if err != nil {
		return parachain.BlockData{}, err
	}

	return parachain.BlockData{
		Block:        block,
		CompactProof: compactProofs,
	}, nil
}
