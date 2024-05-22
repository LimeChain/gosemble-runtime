package decoder

var (
	leafVariant = variant{
		bits: 0b0100_0000,
		mask: 0b1100_0000,
	}
	branchVariant = variant{
		bits: 0b1000_0000,
		mask: 0b1100_0000,
	}
	branchWithValueVariant = variant{
		bits: 0b1100_0000,
		mask: 0b1100_0000,
	}
	leafWithHashedValueVariant = variant{
		bits: 0b0010_0000,
		mask: 0b1110_0000,
	}
	branchWithHashedValueVariant = variant{
		bits: 0b0001_0000,
		mask: 0b1111_0000,
	}
	emptyVariant = variant{
		bits: 0b0000_0000,
		mask: 0b1111_1111,
	}
	compactEncodingVariant = variant{
		bits: 0b0000_0001,
		mask: 0b1111_1111,
	}
	invalidVariant = variant{
		bits: 0b0000_0000,
		mask: 0b0000_0000,
	}
)

// variantsOrderedByBitMask is an array of all variants sorted
// in ascending order by the number of LHS set bits each variant mask has.
// See https://spec.polkadot.network/#defn-node-header
// WARNING: DO NOT MUTATE.
// This array is defined at global scope for performance
// reasons only, instead of having it locally defined in
// the decodeHeaderByte function below.
// For 7 variants, the performance is improved by ~20%.
var variantsOrderedByBitMask = [...]variant{
	leafVariant,                  // mask 1100_0000
	branchVariant,                // mask 1100_0000
	branchWithValueVariant,       // mask 1100_0000
	leafWithHashedValueVariant,   // mask 1110_0000
	branchWithHashedValueVariant, // mask 1111_0000
	emptyVariant,                 // mask 1111_1111
	compactEncodingVariant,       // mask 1111_1111
}

type variant struct {
	bits byte
	mask byte
}

// partialKeyLengthHeaderMask returns the partial key length
// header bit mask corresponding to the variant header bit mask.
// For example for the leaf variant with variant mask 1100_0000,
// the partial key length header mask returned is 0011_1111.
func (v variant) partialKeyLengthHeaderMask() byte {
	return ^v.mask
}
