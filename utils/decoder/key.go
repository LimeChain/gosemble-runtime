package decoder

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ChainSafe/gossamer/pkg/trie/codec"
)

const maxPartialKeyLength = ^uint16(0)

var ErrReaderMismatchCount = errors.New("read unexpected number of bytes from reader")

// decodeKey decodes a key from a reader.
func decodeKey(reader *bytes.Buffer, partialKeyLength uint16) (b []byte, err error) {
	if partialKeyLength == 0 {
		return []byte{}, nil
	}

	key := make([]byte, partialKeyLength/2+partialKeyLength%2)
	n, err := reader.Read(key)
	if err != nil {
		return nil, fmt.Errorf("reading from reader: %w", err)
	} else if n != len(key) {
		return nil, fmt.Errorf("%w: read %d bytes instead of expected %d bytes",
			ErrReaderMismatchCount, n, len(key))
	}

	return codec.KeyLEToNibbles(key)[partialKeyLength%2:], nil
}
