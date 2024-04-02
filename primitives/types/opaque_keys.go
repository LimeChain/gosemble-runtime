package types

import sc "github.com/LimeChain/goscale"

type OpaqueKeys struct {
	keys []SessionKey
}

func NewOpaqueKeys(keys []SessionKey) OpaqueKeys {
	return OpaqueKeys{keys: keys}
}

func (ok OpaqueKeys) KeyTypeIds() []sc.FixedSequence[sc.U8] {
	var keyTypeIds []sc.FixedSequence[sc.U8]

	for _, opaqueKey := range ok.keys {
		keyTypeIds = append(keyTypeIds, opaqueKey.TypeId)
	}

	return keyTypeIds
}
