package parachain

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/ChainSafe/gossamer/lib/runtime/storage"
	"github.com/ChainSafe/gossamer/pkg/trie/db"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/aura_ext"
	"github.com/LimeChain/gosemble/frame/parachain_system"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/parachain"
	"github.com/LimeChain/gosemble/primitives/pvf"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "Parachain"
	apiVersion    = 1
)

// Module implements the Parachain `validate_block` Runtime API function
type Module struct {
	parachainSystem parachain_system.Module
	blockExecutor   aura_ext.BlockExecutor
	runtimeDecoder  types.RuntimeDecoder
	hostEnvironment pvf.HostEnvironment
	hashing         io.Hashing
	logger          log.RuntimeLogger
	memUtils        utils.WasmMemoryTranslator
}

func New(
	parachainSystem parachain_system.Module,
	blockExecutor aura_ext.BlockExecutor,
	runtimeDecoder types.RuntimeDecoder,
	hostEnvironment pvf.HostEnvironment,
	logger log.RuntimeLogger) Module {
	return Module{
		parachainSystem: parachainSystem,
		blockExecutor:   blockExecutor,
		runtimeDecoder:  runtimeDecoder,
		hostEnvironment: hostEnvironment,
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

	blockData, err := m.runtimeDecoder.DecodeParachainBlockData(validationData.BlockData)
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

	parachainInherentData, err := m.extractParachainInherentData(blockData.Block)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	err = validateValidationData(
		parachainInherentData.ValidationData,
		validationData.RelayParentBlockNumber,
		validationData.RelayParentStorageRoot,
		validationData.ParentHead)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	database, err := db.NewMemoryDBFromProof(blockData.CompactProof.ToBytes())
	if err != nil {
		m.logger.Critical(err.Error())
	}
	trie, err := parachain.BuildTrie(parentHeader.StateRoot.Bytes(), database)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	trieState := storage.NewTrieState(trie)
	m.hostEnvironment.SetTrieState(trieState)

	err = m.blockExecutor.ExecuteBlock(blockData.Block)
	if err != nil {
		m.logger.Critical(err.Error())
	}

	collationInfo, err := m.parachainSystem.CollectCollationInfo(blockData.Block.Header())
	if err != nil {
		m.logger.Critical(err.Error())
	}

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

func (m Module) extractParachainInherentData(block primitives.Block) (parachain.InherentData, error) {
	for _, extrinsic := range block.Extrinsics() {
		if extrinsic.IsSigned() {
			continue
		}
		call := extrinsic.Function()
		if call.ModuleIndex() == m.parachainSystem.GetIndex() && call.FunctionIndex() == parachain_system.FunctionSetValidationData {
			parachainInherentData, ok := call.Args()[0].(parachain.InherentData)
			if !ok {
				return parachain.InherentData{}, errors.New("cannot cast to ParachainInherentData")
			}

			return parachainInherentData, nil
		}
	}

	return parachain.InherentData{}, errors.New("not found")
}

func validateValidationData(validationData parachain.PersistedValidationData, relayChainBlockNumber sc.U32, relayParentStorageRoot primitives.H256, parentHead sc.Sequence[sc.U8]) error {
	if !reflect.DeepEqual(validationData.ParentHead, parentHead) {
		return errors.New("parent head doesn't match")
	}
	if !reflect.DeepEqual(validationData.RelayParentNumber, relayChainBlockNumber) {
		return errors.New("relay parent number doesn't match")
	}

	if !reflect.DeepEqual(validationData.RelayParentStorageRoot, relayParentStorageRoot) {
		return errors.New("relay parent storage root doesn't match")
	}

	return nil
}
