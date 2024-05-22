package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type AbridgedHostConfiguration struct {
	MaxCodeSize                     sc.U32
	MaxHeadDataSize                 sc.U32
	MaxUpwardQueueCount             sc.U32
	MaxUpwardQueueSize              sc.U32
	MaxUpwardMessageSize            sc.U32
	MaxUpwardMessageNumPerCandidate sc.U32
	MaxHrmpMessageNumPerCandidate   sc.U32
	ValidationUpgradeCooldown       sc.U32
	ValidationUpgradeDelay          sc.U32
	AsyncBackingParams              AsyncBackingParams
}

func (ahc AbridgedHostConfiguration) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		ahc.MaxCodeSize,
		ahc.MaxHeadDataSize,
		ahc.MaxUpwardQueueCount,
		ahc.MaxUpwardQueueSize,
		ahc.MaxUpwardMessageSize,
		ahc.MaxUpwardMessageNumPerCandidate,
		ahc.MaxHrmpMessageNumPerCandidate,
		ahc.ValidationUpgradeCooldown,
		ahc.ValidationUpgradeDelay,
		ahc.AsyncBackingParams)
}

func DecodeAbridgeHostConfiguration(buffer *bytes.Buffer) (AbridgedHostConfiguration, error) {
	maxCodeSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	maxHeadDataSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	maxUpwardQueueCount, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	maxUpwardQueueSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	maxUpwardMessageSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	maxUpwardMessageNumPerCandidate, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	maxHrmpMessageNumPerCandidate, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	validationUpgradeCooldown, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	validationUpgradeDelay, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	abp, err := DecodeAsyncBackingParams(buffer)
	if err != nil {
		return AbridgedHostConfiguration{}, err
	}

	return AbridgedHostConfiguration{
		MaxCodeSize:                     maxCodeSize,
		MaxHeadDataSize:                 maxHeadDataSize,
		MaxUpwardQueueCount:             maxUpwardQueueCount,
		MaxUpwardQueueSize:              maxUpwardQueueSize,
		MaxUpwardMessageSize:            maxUpwardMessageSize,
		MaxUpwardMessageNumPerCandidate: maxUpwardMessageNumPerCandidate,
		MaxHrmpMessageNumPerCandidate:   maxHrmpMessageNumPerCandidate,
		ValidationUpgradeCooldown:       validationUpgradeCooldown,
		ValidationUpgradeDelay:          validationUpgradeDelay,
		AsyncBackingParams:              abp,
	}, nil
}

func (ahc AbridgedHostConfiguration) Bytes() []byte {
	return sc.EncodedBytes(ahc)
}
