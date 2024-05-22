package parachain_system

import (
	"bytes"
	"errors"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callSetValidationData struct {
	primitives.Callable
	module Module
}

func newCallSetValidationData(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	return callSetValidationData{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(),
		},
		module: module,
	}
}

func newCallSetValidationDataWithArgs(moduleId sc.U8, functionId sc.U8, args sc.VaryingData) primitives.Call {
	call := callSetValidationData{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  args,
		},
	}

	return call
}

func (c callSetValidationData) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	data, err := DecodeParachainInherentData(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(data)
	return c, nil
}
func (c callSetValidationData) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSetValidationData) Bytes() []byte { return c.Callable.Bytes() }

func (c callSetValidationData) ModuleIndex() sc.U8 { return c.Callable.ModuleIndex() }

func (c callSetValidationData) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callSetValidationData) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callSetValidationData) BaseWeight() primitives.Weight {
	return primitives.WeightZero()
}

func (_ callSetValidationData) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callSetValidationData) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassMandatory()
}

func (_ callSetValidationData) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callSetValidationData) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	data, ok := args[0].(ParachainInherentData)
	if !ok {
		return primitives.PostDispatchInfo{}, errors.New("couldn't dispatch call set validation data value")
	}

	return c.setValidationData(origin, data)
}

func (c callSetValidationData) setValidationData(origin primitives.RuntimeOrigin, data ParachainInherentData) (primitives.PostDispatchInfo, error) {
	if !origin.IsNoneOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	if c.module.storage.ValidationData.Exists() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("validation data must be updated only once in a block.")
	}

	totalWeight := primitives.WeightZero()

	lastRelayChainBlockNumber, err := c.module.storage.LastRelayChainBlockNumber.Get()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	c.module.config.CheckAssociatedRelayNumber.CheckAssociatedRelayNumber(data.ValidationData.RelayParentNumber, lastRelayChainBlockNumber)

	parachainId, err := c.module.config.SelfParaId.StorageParaId()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	c.module.logger.Infof("before relay chain state proof init")

	trie, err := parachain.NewRelayChainStateProof(parachainId, data.ValidationData.RelayParentStorageRoot, data.RelayChainState, io.NewHashing())
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	c.module.logger.Infof("after relay chain state proof init")

	consensusHookWeight, capacity, err := c.module.config.ConsensusHook.OnStateProof(trie)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	totalWeight = totalWeight.Add(consensusHookWeight)
	totalWeight = totalWeight.Add(c.module.MaybeDropIncludedAncestors(trie, capacity))

	c.module.config.systemModule.
		DepositLog(
			parachain.NewDigestRelayParentStorageRoot(
				data.ValidationData.RelayParentStorageRoot,
				data.ValidationData.RelayParentNumber,
			),
		)

	// initialization logic: we know that this runs exactly once every block,
	// which means we can put the initialization logic here to remove the
	// sequencing problem.
	upgradeGoAheadSignal, err := trie.ReadUpgradeGoAheadSignal()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	aggr, err := c.module.storage.AggregatedUnincludedSegment.Get()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	if aggr.ConsumedGoAheadSignal.HasValue {
		if aggr.ConsumedGoAheadSignal.Value != upgradeGoAheadSignal.Value {
			c.module.logger.Critical("Mismatching Go Ahead signals")
		}
	}

	if upgradeGoAheadSignal.HasValue {
		switch upgradeGoAheadSignal.Value {
		case parachain.UpgradeGoAheadGoAhead:
			if !c.module.storage.PendingValidationCode.Exists() {
				c.module.logger.Critical("No new validation function found in storage, GoAhead signal is not expected.")
			}
			validationCode, err := c.module.storage.PendingValidationCode.Take()
			if err != nil {
				c.module.logger.Infof(err.Error())
				return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
			}
			c.module.config.systemModule.UpdateCodeInStorage(validationCode)
			c.module.config.systemModule.DepositEvent(newEventValidationFunctionApplied(c.ModuleId, data.ValidationData.RelayParentNumber))
		case parachain.UpgradeGoAheadAbort:
			c.module.storage.PendingValidationCode.Clear()
			c.module.config.systemModule.DepositEvent(newEventValidationFunctionDiscarded(c.ModuleId))
		}
	}

	restrictionSignal, err := trie.ReadRestrictionSignal()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	c.module.storage.UpgradeRestrictionSignal.Put(restrictionSignal)
	c.module.storage.UpgradeGoAhead.Put(upgradeGoAheadSignal)

	hostConfig, err := trie.ReadAbridgedHostConfiguration()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	relevantMessagingState, err := trie.ReadMessagingStateSnapshot(hostConfig)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	c.module.storage.ValidationData.Put(data.ValidationData)
	c.module.storage.RelayStateProof.Put(data.RelayChainState)
	c.module.storage.RelevantMessagingState.Put(relevantMessagingState)
	c.module.storage.HostConfiguration.Put(hostConfig)

	// TODO:
	//total_weight.saturating_accrue(Self::enqueue_inbound_downward_messages(
	//	relevant_messaging_state.dmq_mqc_head,
	//	downward_messages,
	//));
	//total_weight.saturating_accrue(Self::enqueue_inbound_horizontal_messages(
	//	&relevant_messaging_state.ingress_channels,
	//	horizontal_messages,
	//	vfp.relay_parent_number,
	//));

	return primitives.PostDispatchInfo{
		ActualWeight: sc.NewOption[primitives.Weight](totalWeight),
		PaysFee:      primitives.PaysNo,
	}, nil
}

func (c callSetValidationData) Docs() string {
	return "Set the current validation data. This should be invoked exactly once per block. " +
		"It will panic at the finalisation if the call was not invoked. " +
		"The dispatch origin for this call must be `Inherent`. " +
		"As a side effect, this function upgrades the current validation function if the appropriate time has come."
}
