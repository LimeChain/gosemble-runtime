package session

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	keySession            = []byte("Session")
	keyValidators         = []byte("Validators")
	keyCurrentIndex       = []byte("CurrentIndex")
	keyQueuedChanged      = []byte("QueuedChanged")
	keyQueuedKeys         = []byte("QueuedKeys")
	keyDisabledValidators = []byte("DisabledValidators")
	keyNextKeys           = []byte("NextKeys")
	keyKeyOwner           = []byte("KeyOwner")
)

type storage struct {
	Validators         support.StorageValue[sc.Sequence[primitives.AccountId]]
	CurrentIndex       support.StorageValue[sc.U32]
	QueueChanged       support.StorageValue[sc.Bool]
	QueuedKeys         support.StorageValue[sc.Sequence[queuedKey]]
	DisabledValidators support.StorageValue[sc.Sequence[sc.U32]]
	NextKeys           support.StorageMap[primitives.AccountId, sc.FixedSequence[primitives.Sr25519PublicKey]]
	KeyOwner           support.StorageMap[primitives.SessionKey, primitives.AccountId]
}

func newStorage(s io.Storage, m Module) *storage {
	hashing := io.NewHashing()

	return &storage{
		Validators:         support.NewHashStorageValue(s, keySession, keyValidators, primitives.DecodeSequenceAccountId),
		CurrentIndex:       support.NewHashStorageValue(s, keySession, keyCurrentIndex, sc.DecodeU32),
		QueueChanged:       support.NewHashStorageValue(s, keySession, keyQueuedChanged, sc.DecodeBool),
		QueuedKeys:         support.NewHashStorageValue(s, keySession, keyQueuedKeys, DecodeQueuedKeys),
		DisabledValidators: support.NewHashStorageValue(s, keySession, keyDisabledValidators, sc.DecodeSequence[sc.U32]),
		NextKeys:           support.NewHashStorageMap[primitives.AccountId, sc.FixedSequence[primitives.Sr25519PublicKey]](s, keySession, keyNextKeys, hashing.Twox64, m.handler.DecodeKeys),
		KeyOwner:           support.NewHashStorageMap[primitives.SessionKey, primitives.AccountId](s, keySession, keyKeyOwner, hashing.Twox64, primitives.DecodeAccountId),
	}
}
