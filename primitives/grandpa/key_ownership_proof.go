package grandpa

import sc "github.com/LimeChain/goscale"

type KeyOwnerProof interface {
	sc.Encodable
	Session() sc.U32
	ValidatorCount() sc.U32
}

// An opaque type used to represent the key ownership proof at the runtime API
// boundary. The inner value is an encoded representation of the actual key
// ownership proof which will be parameterized when defining the runtime. At
// the runtime API boundary this type is unknown and as such we keep this
// opaque representation, implementors of the runtime API will have to make
// sure that all usages of `OpaqueKeyOwnershipProof` refer to the same type.
type OpaqueKeyOwnershipProof = sc.Sequence[sc.U8]
