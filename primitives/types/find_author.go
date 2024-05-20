package types

import sc "github.com/LimeChain/goscale"

type FindAuthor[T sc.Encodable] interface {
	FindAuthor(digests sc.Sequence[DigestPreRuntime]) (sc.Option[T], error)
}
