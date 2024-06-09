package parachain_system

import sc "github.com/LimeChain/goscale"

const (
	ErrorOverlappingUpgrades sc.U8 = iota
	ErrorProhibitedByPolkadot
	ErrorTooBig
	ErrorValidationDataNotAvailable
	ErrorHostConfigurationNotAvailable
	ErrorNotScheduled
	ErrorNothingAuthorized
	ErrorUnauthorized
)

const (
	InherentErrorInvalid sc.U8 = iota
)

type InherentError struct {
	sc.VaryingData
}

func NewInherentErrorInvalid() InherentError {
	return InherentError{sc.NewVaryingData(InherentErrorInvalid)}
}

func (ie InherentError) IsFatal() sc.Bool {
	switch ie.VaryingData[0] {
	case InherentErrorInvalid:
		return true
	default:
		return false
	}
}

func (ie InherentError) Error() string {
	switch ie.VaryingData[0] {
	case InherentErrorInvalid:
		return "invalid inherent check for parachain system module"
	default:
		return "not a valid inherent error type"
	}
}
