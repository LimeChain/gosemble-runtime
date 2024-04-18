package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type UpwardMessage = sc.Sequence[sc.U8]

func DecodeUpwardMessages(buffer *bytes.Buffer) (sc.Sequence[UpwardMessage], error) {
	return sc.DecodeSequenceWith(buffer, sc.DecodeSequence[sc.U8])
}
