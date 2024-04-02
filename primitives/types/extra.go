package types

import (
	"bytes"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
)

type SignedExtra interface {
	sc.Encodable

	Decode(buffer *bytes.Buffer)
	DeepCopy() SignedExtra

	AdditionalSigned() (AdditionalSigned, error)
	Validate(who AccountId, call Call, info *DispatchInfo, length sc.Compact) (ValidTransaction, error)
	ValidateUnsigned(call Call, info *DispatchInfo, length sc.Compact) (ValidTransaction, error)
	PreDispatch(who AccountId, call Call, info *DispatchInfo, length sc.Compact) (sc.Sequence[Pre], error)
	PreDispatchUnsigned(call Call, info *DispatchInfo, length sc.Compact) error
	PostDispatch(pre sc.Option[sc.Sequence[Pre]], info *DispatchInfo, postInfo *PostDispatchInfo, length sc.Compact, dispatchErr error) error

	Metadata() sc.Sequence[MetadataSignedExtension]
}

// signedExtra contains an array of SignedExtension, iterated through during extrinsic execution.
type signedExtra struct {
	extras      []SignedExtension
	mdGenerator *MetadataTypeGenerator
}

func NewSignedExtra(checks []SignedExtension, mdGenerator *MetadataTypeGenerator) SignedExtra {
	return signedExtra{
		extras:      checks,
		mdGenerator: mdGenerator,
	}
}

func (e signedExtra) Encode(buffer *bytes.Buffer) error {
	for _, extra := range e.extras {
		err := extra.Encode(buffer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e signedExtra) Bytes() []byte {
	return sc.EncodedBytes(e)
}

func (e signedExtra) Decode(buffer *bytes.Buffer) {
	for _, extra := range e.extras {
		extra.Decode(buffer)
	}
}

func (e signedExtra) DeepCopy() SignedExtra {
	var extras []SignedExtension
	for _, extra := range e.extras {
		extras = append(extras, extra.DeepCopy())
	}

	return signedExtra{
		extras:      extras,
		mdGenerator: e.mdGenerator,
	}
}

func (e signedExtra) AdditionalSigned() (AdditionalSigned, error) {
	result := AdditionalSigned{}

	for _, extra := range e.extras {
		ok, err := extra.AdditionalSigned()
		if err != nil {
			return nil, err
		}
		result = append(result, ok...)
	}

	return result, nil
}

func (e signedExtra) Validate(who AccountId, call Call, info *DispatchInfo, length sc.Compact) (ValidTransaction, error) {
	valid := DefaultValidTransaction()

	for _, extra := range e.extras {
		v, err := extra.Validate(who, call, info, length)
		if err != nil {
			return ValidTransaction{}, err
		}
		valid = valid.CombineWith(v)
	}

	return valid, nil
}

func (e signedExtra) ValidateUnsigned(call Call, info *DispatchInfo, length sc.Compact) (ValidTransaction, error) {
	valid := DefaultValidTransaction()

	for _, extra := range e.extras {
		v, err := extra.ValidateUnsigned(call, info, length)
		if err != nil {
			return ValidTransaction{}, err
		}
		valid = valid.CombineWith(v)
	}

	return valid, nil
}

func (e signedExtra) PreDispatch(who AccountId, call Call, info *DispatchInfo, length sc.Compact) (sc.Sequence[Pre], error) {
	pre := sc.Sequence[Pre]{}

	for _, extra := range e.extras {
		p, err := extra.PreDispatch(who, call, info, length)
		if err != nil {
			return nil, err
		}

		pre = append(pre, p)
	}

	return pre, nil
}

func (e signedExtra) PreDispatchUnsigned(call Call, info *DispatchInfo, length sc.Compact) error {
	for _, extra := range e.extras {
		err := extra.PreDispatchUnsigned(call, info, length)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e signedExtra) PostDispatch(pre sc.Option[sc.Sequence[Pre]], info *DispatchInfo, postInfo *PostDispatchInfo, length sc.Compact, dispatchErr error) error {
	if pre.HasValue {
		preValue := pre.Value
		for i, extra := range e.extras {
			err := extra.PostDispatch(sc.NewOption[Pre](preValue[i]), info, postInfo, length, dispatchErr)
			if err != nil {
				return err
			}
		}
	} else {
		for _, extra := range e.extras {

			err := extra.PostDispatch(sc.NewOption[Pre](nil), info, postInfo, length, dispatchErr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e signedExtra) Metadata() sc.Sequence[MetadataSignedExtension] {
	ids := sc.Sequence[sc.Compact]{}
	signedExtensions := sc.Sequence[MetadataSignedExtension]{}

	for _, extra := range e.extras {
		extraMetadataId := e.mdGenerator.BuildExtraMetadata(reflect.ValueOf(extra), &signedExtensions)
		ids = append(ids, sc.ToCompact(extraMetadataId))
	}

	signedExtraType := NewMetadataType(metadata.SignedExtra, "SignedExtra", NewMetadataTypeDefinitionTuple(ids))
	e.mdGenerator.AppendMetadataTypes(sc.Sequence[MetadataType]{signedExtraType})

	return signedExtensions
}
