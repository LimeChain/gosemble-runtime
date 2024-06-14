package parachain

import sc "github.com/LimeChain/goscale"

const (
	UnincludedSegmentCapacityInnerExpectParentIncluded sc.U8 = iota
	UnincludedSegmentCapacityInnerValue
)

func NewUnincludedSegmentCapacityInnerExpectParentIncluded() UnincludedSegmentCapacity {
	return UnincludedSegmentCapacity{sc.NewVaryingData(UnincludedSegmentCapacityInnerExpectParentIncluded)}
}

func NewUnincludedSegmentCapacityValue(value sc.U32) UnincludedSegmentCapacity {
	return UnincludedSegmentCapacity{sc.NewVaryingData(UnincludedSegmentCapacityInnerValue, value)}
}

type UnincludedSegmentCapacity struct {
	sc.VaryingData
}

func (usc UnincludedSegmentCapacity) Get() sc.U32 {
	switch usc.VaryingData[0] {
	case UnincludedSegmentCapacityInnerExpectParentIncluded:
		return 1
	case UnincludedSegmentCapacityInnerValue:
		return usc.VaryingData[1].(sc.U32)
	default:
		panic("invalid UnincludedSegmentCapacity value")
	}
}

func (usc UnincludedSegmentCapacity) IsExpectingIncludedParent() bool {
	switch usc.VaryingData[0] {
	case UnincludedSegmentCapacityInnerExpectParentIncluded:
		return true
	case UnincludedSegmentCapacityInnerValue:
		return false
	default:
		panic("invalid UnincludedSegmentCapacity value")
	}
}
