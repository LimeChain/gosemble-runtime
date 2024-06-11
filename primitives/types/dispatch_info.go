package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

// DispatchInfo A bundle of static information collected from the `#[pallet::weight]` attributes.
type DispatchInfo struct {
	// Weight of this transaction.
	Weight Weight

	// Class of this transaction.
	Class DispatchClass

	// Does this transaction pay fees.
	PaysFee Pays
}

func (di DispatchInfo) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		di.Weight,
		di.Class,
		di.PaysFee,
	)
}

func DecodeDispatchInfo(buffer *bytes.Buffer) (DispatchInfo, error) {
	di := DispatchInfo{}
	weight, err := DecodeWeight(buffer)
	if err != nil {
		return DispatchInfo{}, err
	}
	di.Weight = weight
	class, err := DecodeDispatchClass(buffer)
	if err != nil {
		return DispatchInfo{}, err
	}
	di.Class = class
	paysFee, err := DecodePays(buffer)
	if err != nil {
		return DispatchInfo{}, err
	}
	di.PaysFee = paysFee
	return di, nil
}

func (di DispatchInfo) Bytes() []byte {
	return sc.EncodedBytes(di)
}

func (di DispatchInfo) IsMandatory() (sc.Bool, error) {
	return di.Class.Is(DispatchClassMandatory)
}

// GetDispatchInfo returns the DispatchInfo of the given call.
// Uses call's BaseWeight to calculate all the information in DispatchInfo
func GetDispatchInfo(call Call) DispatchInfo {
	baseWeight := call.BaseWeight()

	return DispatchInfo{
		Weight:  call.WeighData(baseWeight),
		Class:   call.ClassifyDispatch(baseWeight),
		PaysFee: call.PaysFee(baseWeight),
	}
}
