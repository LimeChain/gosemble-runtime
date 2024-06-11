package grandpa

import (
	"bytes"
	"errors"
	"fmt"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/authorship"
	"github.com/LimeChain/gosemble/frame/session_historical"
	grandpatypes "github.com/LimeChain/gosemble/primitives/grandpa"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/session"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// A round number and set id which point on the time of an offence.
type TimeSlot struct {
	// Grandpa Set ID.
	SetId sc.U64
	// Round number.
	Round sc.U64
}

// GRANDPA equivocation offence report.
type EquivocationOffence struct {
	// Time slot at which this inEquivocationOffencecident happened.
	TimeSlot TimeSlot
	// The session index in which the incident happened.
	SessionIndex sc.U32
	// The size of the validator set at the time of the offence.
	ValidatorSetCount sc.U32
	// The authority which produced this equivocation.
	Offender primitives.AccountId
}

type EquivocationReportSystem struct {
	grandpaModule           Module
	authorshipModule        authorship.Module
	sessionHistoricalModule session_historical.Module
	logger                  log.RuntimeLogger
}

func NewEquivocationReportSystem(sessionHistoricalModule session_historical.Module, authorshipModule authorship.Module, logger log.RuntimeLogger) EquivocationReportSystem {
	return EquivocationReportSystem{
		authorshipModule:        authorshipModule,
		sessionHistoricalModule: sessionHistoricalModule,
		logger:                  logger,
	}
}

func (e EquivocationReportSystem) SetModule(grandpaModule Module) {
	e.grandpaModule = grandpaModule
}

func (e EquivocationReportSystem) PublishEvidence(equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof) error {
	call := callReportEquivocationUnsigned{
		Callable: primitives.Callable{
			ModuleId:   e.grandpaModule.GetIndex(),
			FunctionId: functionReportEquivocationUnsignedIndex,
			Arguments:  sc.NewVaryingData(equivocationProof, keyOwnerProof),
		},
	}

	// TODO: implement it as part of the system module
	// res := SubmitTransaction::<T, Call<T>>::submit_unsigned_transaction(call.into())
	xt := types.NewUnsignedUncheckedExtrinsic(call)
	buffer := &bytes.Buffer{}
	err := xt.Encode(buffer)
	if err != nil {
		return errors.New(fmt.Sprintf("Error submitting equivocation report: %s", err))
	}
	io.NewOffchain().SubmitTransaction(buffer.Bytes())

	e.logger.Warn(fmt.Sprint("Submitted equivocation report"))

	return nil
}

func (e EquivocationReportSystem) ProcessEvidence(reporterAccount sc.Option[primitives.AccountId], equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof,
) error {
	var reporter primitives.AccountId
	if reporterAccount.HasValue {
		reporter = reporterAccount.Value
	} else {
		author, err := e.authorshipModule.Author()
		if err != nil {
			return err
		}
		if author.HasValue {
			reporter = author.Value
		}
	}

	// TODO: handle errors

	offender, _ := equivocationProof.Offender()

	// We check the equivocation within the context of its set id (and
	// associated session) and round. We also need to know the validator
	// set count when the offence since it is required to calculate the
	// slash amount.
	setId := equivocationProof.SetId
	round, _ := equivocationProof.Round()
	sessionIndex := keyOwnerProof.Session()
	validatorSetCount := keyOwnerProof.ValidatorCount()

	// Validate equivocation proof (check votes are different and signatures are valid).
	if !grandpatypes.CheckEquivocationProof(equivocationProof) {
		return NewDispatchErrorInvalidEquivocationProof(e.grandpaModule.GetIndex())
	}

	// Validate the key ownership proof extracting the id of the offender.
	offenderId := e.sessionHistoricalModule.CheckProof(e.grandpaModule.KeyTypeId(), offender, keyOwnerProof.(session.MembershipProof))
	if offenderId.HasValue {
		return NewDispatchErrorInvalidKeyOwnershipProof(e.grandpaModule.GetIndex())
	}

	// Fetch the current and previous sets last session index.
	// For genesis set there's no previous set.
	var previousSetIdSessionIndex sc.Option[sc.U32]
	if setId != 0 {
		idx, err := e.grandpaModule.StorageSetIdSessionGet(setId - 1)
		if err != nil {
			return NewDispatchErrorInvalidEquivocationProof(e.grandpaModule.GetIndex())
		}
		previousSetIdSessionIndex = sc.NewOption[sc.U32](idx)
	} else {
		previousSetIdSessionIndex = sc.NewOption[sc.U32](nil)
	}

	setIdSessionIndex, err := e.grandpaModule.StorageSetIdSessionGet(setId)
	if err != nil {
		return NewDispatchErrorInvalidEquivocationProof(e.grandpaModule.GetIndex())
	}

	// Check that the session id for the membership proof is within the
	// bounds of the set id reported in the equivocation.
	var prevCheck bool
	if previousSetIdSessionIndex.HasValue {
		previousIndex := previousSetIdSessionIndex.Value
		prevCheck = sessionIndex <= previousIndex
	} else {
		prevCheck = false
	}

	if sessionIndex > setIdSessionIndex || prevCheck {
		return NewDispatchErrorInvalidEquivocationProof(e.grandpaModule.GetIndex())
	}

	offence := EquivocationOffence{
		TimeSlot:          TimeSlot{setId, round},
		SessionIndex:      sessionIndex,
		Offender:          offender,
		ValidatorSetCount: validatorSetCount,
	}

	// TODO:
	// R::report_offence(reporter.into_iter().collect(), offence).map_err(|_| Error::<T>::DuplicateOffenceReport)?;
	_ = reporter
	_ = offence

	return nil
}
