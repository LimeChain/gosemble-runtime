package executive

import (
	"errors"
	"fmt"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/extrinsic"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	errInvalidParentHash  = errors.New("parent hash should be valid")
	errInvalidDigestNum   = errors.New("number of digest must match the calculated")
	errInvalidDigestItem  = errors.New("digest item must match that calculated")
	errInvalidStorageRoot = errors.New("storage root must match that calculated")
	errInvalidTxTrie      = errors.New("Transaction trie must be valid")
)

type Module interface {
	InitializeBlock(header primitives.Header) error
	ExecuteBlock(block primitives.Block) error
	ApplyExtrinsic(uxt primitives.UncheckedExtrinsic) error
	FinalizeBlock() (primitives.Header, error)
	ValidateTransaction(source primitives.TransactionSource, uxt primitives.UncheckedExtrinsic, blockHash primitives.Blake2bHash) (primitives.ValidTransaction, error)
	OffchainWorker(header primitives.Header) error
}

type module struct {
	system           system.Module
	onRuntimeUpgrade primitives.OnRuntimeUpgrade
	runtimeExtrinsic extrinsic.RuntimeExtrinsic
	hashing          io.Hashing
	logger           log.TraceLogger
}

func New(systemModule system.Module, runtimeExtrinsic extrinsic.RuntimeExtrinsic, onRuntimeUpgrade primitives.OnRuntimeUpgrade, logger log.TraceLogger) Module {
	return module{
		system:           systemModule,
		onRuntimeUpgrade: onRuntimeUpgrade,
		runtimeExtrinsic: runtimeExtrinsic,
		hashing:          io.NewHashing(),
		logger:           logger,
	}
}

// InitializeBlock initialises a block with the given header,
// starting the execution of a particular block.
func (m module) InitializeBlock(header primitives.Header) error {
	m.logger.Trace("init_block")
	m.system.ResetEvents()

	weight := primitives.WeightZero()
	upgrade, err := m.runtimeUpgrade()
	if err != nil {
		return err
	}
	if upgrade {
		weight = weight.SaturatingAdd(m.executeOnRuntimeUpgrade())
	}

	m.system.Initialize(header.Number, header.ParentHash, extractPreRuntimeDigest(header.Digest))

	onInit, err := m.runtimeExtrinsic.OnInitialize(header.Number)
	if err != nil {
		return err
	}

	weight = weight.SaturatingAdd(onInit)
	weight = weight.SaturatingAdd(m.system.BlockWeights().BaseBlock)
	// use in case of dynamic weight calculation
	err = m.system.RegisterExtraWeightUnchecked(weight, primitives.NewDispatchClassMandatory())
	if err != nil {
		return err
	}

	m.system.NoteFinishedInitialize()
	return nil
}

func (m module) ExecuteBlock(block primitives.Block) error {
	m.logger.Tracef("execute_block %v", block.Header().Number)

	err := m.InitializeBlock(block.Header())
	if err != nil {
		return err
	}

	err = m.initialChecks(block)
	if err != nil {
		return err
	}

	// todo: handle err
	m.executeExtrinsicsWithBookKeeping(block)

	header := block.Header()
	err = m.finalChecks(&header)
	if err != nil {
		return err
	}
	return nil
}

// ApplyExtrinsic applies extrinsic outside the block execution function.
//
// This doesn't attempt to validate anything regarding the block, but it builds a list of uxt
// hashes.
func (m module) ApplyExtrinsic(uxt primitives.UncheckedExtrinsic) error {
	encoded := uxt.Bytes()
	encodedLen := sc.ToCompact(len(encoded))

	m.logger.Trace("apply_extrinsic")

	// Verify that the signature is good.
	checked, err := uxt.Check()
	if err != nil {
		return err
	}

	// We don't need to make sure to `note_extrinsic` only after we know it's going to be
	// executed to prevent it from leaking in storage since at this point, it will either
	// execute or panic (and revert storage changes).
	m.system.NoteExtrinsic(encoded)

	// AUDIT: Under no circumstances may this function panic from here onwards.

	// Decode parameters and dispatch
	dispatchInfo := primitives.GetDispatchInfo(checked.Function())
	m.logger.Tracef("get_dispatch_info: weight ref time %d", dispatchInfo.Weight.RefTime)
	unsignedValidator := extrinsic.NewUnsignedValidatorForChecked(m.runtimeExtrinsic)
	res, err := checked.Apply(unsignedValidator, &dispatchInfo, encodedLen)

	if err != nil {
		_, isDispatchErr := err.(primitives.DispatchError)
		if !isDispatchErr {
			return err
		}

		// Mandatory(inherents) are not allowed to fail.
		//
		// The entire block should be discarded if an inherent fails to apply. Otherwise
		// it may open an attack vector.
		if isMendatory, err := dispatchInfo.IsMendatory(); err != nil {
			return err
		} else if isMendatory {
			return primitives.NewTransactionValidityError(primitives.NewInvalidTransactionBadMandatory())
		}
	}

	if err := m.system.NoteAppliedExtrinsic(res, err, dispatchInfo); err != nil {
		return err
	}

	return err
}

func (m module) FinalizeBlock() (primitives.Header, error) {
	m.logger.Trace("finalize_block")
	err := m.system.NoteFinishedExtrinsics()
	if err != nil {
		return primitives.Header{}, err
	}
	blockNumber, err := m.system.StorageBlockNumber()
	if err != nil {
		return primitives.Header{}, err
	}

	err = m.idleAndFinalizeHook(blockNumber)
	if err != nil {
		return primitives.Header{}, err
	}

	return m.system.Finalize()
}

// ValidateTransaction checks a given signed transaction for validity. This doesn't execute any
// side-effects; it merely checks whether the transaction would panic if it were included or
// not.
//
// Changes made to storage should be discarded.
func (m module) ValidateTransaction(source primitives.TransactionSource, uxt primitives.UncheckedExtrinsic, blockHash primitives.Blake2bHash) (primitives.ValidTransaction, error) {
	m.logger.Trace("validate_transaction")
	currentBlockNumber, err := m.system.StorageBlockNumber()
	if err != nil {
		return primitives.ValidTransaction{}, err
	}

	m.system.Initialize(currentBlockNumber+1, blockHash, primitives.Digest{})

	m.logger.Trace("using_encoded")
	encodedLen := sc.ToCompact(len(uxt.Bytes()))

	m.logger.Trace("check")
	checked, errCheck := uxt.Check()
	if errCheck != nil {
		return primitives.ValidTransaction{}, errCheck
	}

	m.logger.Trace("dispatch_info")
	dispatchInfo := primitives.GetDispatchInfo(checked.Function())

	if isMendatory, err := dispatchInfo.IsMendatory(); err != nil {
		return primitives.ValidTransaction{}, err
	} else if isMendatory {
		return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewInvalidTransactionMandatoryValidation())
	}

	m.logger.Trace("validate")
	unsignedValidator := extrinsic.NewUnsignedValidatorForChecked(m.runtimeExtrinsic)
	return checked.Validate(unsignedValidator, source, &dispatchInfo, encodedLen)
}

func (m module) OffchainWorker(header primitives.Header) error {
	m.system.Initialize(header.Number, header.ParentHash, header.Digest)

	hash := m.hashing.Blake256(header.Bytes())
	blockHash, err := primitives.NewBlake2bHash(sc.BytesToSequenceU8(hash)...)
	if err != nil {
		return err
	}

	m.system.StorageBlockHashSet(header.Number, blockHash)

	m.runtimeExtrinsic.OffchainWorker(header.Number)

	return nil
}

func (m module) idleAndFinalizeHook(blockNumber sc.U64) error {
	weight, err := m.system.StorageBlockWeight()
	if err != nil {
		return err
	}

	maxWeight := m.system.BlockWeights().MaxBlock

	total, totalErr := weight.Total()
	if totalErr != nil {
		return totalErr
	}
	remainingWeight := maxWeight.SaturatingSub(total)

	if remainingWeight.AllGt(primitives.WeightZero()) {
		usedWeight := m.runtimeExtrinsic.OnIdle(blockNumber, remainingWeight)
		err = m.system.RegisterExtraWeightUnchecked(usedWeight, primitives.NewDispatchClassMandatory())
		if err != nil {
			return err
		}
	}

	err = m.runtimeExtrinsic.OnFinalize(blockNumber)
	if err != nil {
		return err
	}
	return nil
}

func (m module) executeExtrinsicsWithBookKeeping(block primitives.Block) error {
	for _, ext := range block.Extrinsics() {

		if err := m.ApplyExtrinsic(ext); err != nil {
			return err
		}
	}

	if err := m.system.NoteFinishedExtrinsics(); err != nil {
		return err
	}

	return m.idleAndFinalizeHook(block.Header().Number)
}

func (m module) initialChecks(block primitives.Block) error {
	m.logger.Trace("initial_checks")

	header := block.Header()
	blockNumber := header.Number

	if blockNumber > 0 {
		storageParentHash, err := m.system.StorageBlockHash(blockNumber - 1)
		if err != nil {
			return err
		}

		if !reflect.DeepEqual(storageParentHash, header.ParentHash) {
			return errInvalidParentHash
		}
	}

	inherentsAreFirst := m.runtimeExtrinsic.EnsureInherentsAreFirst(block)
	if inherentsAreFirst >= 0 {
		return fmt.Errorf("invalid inherent position for extrinsic at index [%d]", inherentsAreFirst)
	}
	return nil
}

func (m module) runtimeUpgrade() (sc.Bool, error) {
	last, err := m.system.StorageLastRuntimeUpgrade()
	if err != nil {
		return false, err
	}

	if last.SpecVersion.Number == nil {
		last.SpecVersion = sc.Compact{Number: sc.U32(0)}
	}

	specVersion := last.SpecVersion.Number.(sc.U32)

	if m.system.Version().SpecVersion > specVersion ||
		last.SpecName != m.system.Version().SpecName {

		current := primitives.LastRuntimeUpgradeInfo{
			SpecVersion: sc.Compact{Number: m.system.Version().SpecVersion},
			SpecName:    m.system.Version().SpecName,
		}
		m.system.StorageLastRuntimeUpgradeSet(current)

		return true, nil
	}

	return false, nil
}

func (m module) finalChecks(header *primitives.Header) error {
	newHeader, err := m.system.Finalize()
	if err != nil {
		return err
	}

	if len(header.Digest.Sequence) != len(newHeader.Digest.Sequence) {
		return errInvalidDigestNum
	}

	for i, digest := range header.Digest.Sequence {
		otherDigest := newHeader.Digest.Sequence[i]
		if !reflect.DeepEqual(digest, otherDigest) {
			return errInvalidDigestItem
		}
	}

	if !reflect.DeepEqual(header.StateRoot, newHeader.StateRoot) {
		return errInvalidStorageRoot
	}

	if !reflect.DeepEqual(header.ExtrinsicsRoot, newHeader.ExtrinsicsRoot) {
		return errInvalidTxTrie
	}
	return nil
}

// executeOnRuntimeUpgrade - Execute all `OnRuntimeUpgrade` of this runtime, and return the aggregate weight.
func (m module) executeOnRuntimeUpgrade() primitives.Weight {
	weight := m.onRuntimeUpgrade.OnRuntimeUpgrade()

	return weight.SaturatingAdd(m.runtimeExtrinsic.OnRuntimeUpgrade())
}

func extractPreRuntimeDigest(digest primitives.Digest) primitives.Digest {
	return digest.OnlyPreRuntimes()
}
