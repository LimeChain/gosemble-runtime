package parachain

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

func Test_MessagingStateSnapshot_Encode(t *testing.T) {
	a := MessagingStateSnapshot{
		DmqMqcHead:                          primitives.H256{FixedSequence: constants.ZeroAccountId.FixedSequence},
		RelayDispatchQueueRemainingCapacity: RelayDispatchQueueRemainingCapacity{},
		IngressChannels:                     nil,
		EgressChannels:                      nil,
	}
	bytesMss := a.Bytes()
	fmt.Println(len(bytesMss))

	mss, err := DecodeMessagingStateSnapshot(bytes.NewBuffer(bytesMss))
	if err != nil {
		panic(err)
	}

	fmt.Println(mss)
}
