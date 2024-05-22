package parachain

import sc "github.com/LimeChain/goscale"

type ReadEntryError = sc.U8

const (
	ReadEntryErrorProof ReadEntryError = iota
	ReadEntryErrorDecode
	ReadEntryErrorAbsent
)
