package types

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// CheckedExtrinsic is the definition of something that the external world might want to say; its
// existence implies that it has been checked and is good, particularly
// in regard to the signature.
type checkedExtrinsic struct {
	// Who this purports to be from and the number of extrinsics have come before
	// from the same signer, if anyone (note this is not a signature).
	signer        sc.Option[primitives.AccountId]
	function      primitives.Call
	extra         primitives.SignedExtra
	transactional support.Transactional[primitives.PostDispatchInfo]
}

func NewCheckedExtrinsic(signer sc.Option[primitives.AccountId], function primitives.Call, extra primitives.SignedExtra, storage io.Storage, transactionBroker io.TransactionBroker, logger log.RuntimeLogger) primitives.CheckedExtrinsic {
	return checkedExtrinsic{
		signer:        signer,
		function:      function,
		extra:         extra,
		transactional: support.NewTransactional[primitives.PostDispatchInfo](storage, transactionBroker, logger),
	}
}

func (c checkedExtrinsic) Function() primitives.Call {
	return c.function
}

func (c checkedExtrinsic) Apply(validator primitives.UnsignedValidator, info *primitives.DispatchInfo, length sc.Compact) (primitives.PostDispatchInfo, error) {
	var (
		maybeWho sc.Option[primitives.AccountId]
		maybePre sc.Option[sc.Sequence[primitives.Pre]]
	)

	if c.signer.HasValue {
		id := c.signer.Value
		pre, err := c.extra.PreDispatch(id, c.function, info, length)
		if err != nil {
			return primitives.PostDispatchInfo{}, err
		}
		maybeWho, maybePre = sc.NewOption[primitives.AccountId](id), sc.NewOption[sc.Sequence[primitives.Pre]](pre)
	} else {
		// Do any pre-flight stuff for an unsigned transaction.
		//
		// Note this function by default delegates to `ValidateUnsigned`, so that
		// all checks performed for the transaction queue are also performed during
		// the dispatch phase (applying the extrinsic).
		//
		// If you ever override this function, you need to make sure to always
		// perform the same validation as in `ValidateUnsigned`.
		err := c.extra.PreDispatchUnsigned(c.function, info, length)
		if err != nil {
			return primitives.PostDispatchInfo{}, err
		}

		_, err = validator.PreDispatch(c.function)
		if err != nil {
			return primitives.PostDispatchInfo{}, err
		}

		maybeWho, maybePre = sc.NewOption[primitives.AccountId](nil), sc.NewOption[sc.Sequence[primitives.Pre]](nil)
	}

	// TODO: revise if the error handling is correct
	postInfo, err := c.transactional.WithStorageLayer(
		func() (primitives.PostDispatchInfo, error) {
			return c.dispatch(maybeWho)
		},
	)

	if err := c.extra.PostDispatch(maybePre, info, &postInfo, length, err); err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	return postInfo, err
}

func (c checkedExtrinsic) Validate(validator primitives.UnsignedValidator, source primitives.TransactionSource, info *primitives.DispatchInfo, length sc.Compact) (primitives.ValidTransaction, error) {
	if c.signer.HasValue {
		id := c.signer.Value
		return c.extra.Validate(id, c.function, info, length)
	}

	valid, err := c.extra.ValidateUnsigned(c.function, info, length)
	if err != nil {
		return primitives.ValidTransaction{}, err
	}

	unsignedValidation, err := validator.ValidateUnsigned(source, c.function)
	if err != nil {
		return primitives.ValidTransaction{}, err
	}

	return valid.CombineWith(unsignedValidation), nil
}

func (c checkedExtrinsic) dispatch(maybeWho sc.Option[primitives.AccountId]) (primitives.PostDispatchInfo, error) {
	return c.function.Dispatch(primitives.RawOriginFrom(maybeWho), c.function.Args())
}
