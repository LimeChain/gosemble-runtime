package session

import (
	"errors"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	errAuthorNotFound         = errors.New("no author found")
	errAuthorIndexOutOfBounds = errors.New("author index out of bounds")
)

// Wraps the author-scraping logic for consensus engines that can recover
// the canonical index of an author. This then transforms it into the
// registering account-ID of that session key index.
type FindAccountFromAuthorIndex struct {
	sessionModule Module
	authorFinder  primitives.FindAuthor[sc.U32]
}

func NewFindAccountFromAuthorIndex(sessionModule Module, authorFinder primitives.FindAuthor[sc.U32]) FindAccountFromAuthorIndex {
	return FindAccountFromAuthorIndex{
		sessionModule: sessionModule,
		authorFinder:  authorFinder,
	}
}

func (f FindAccountFromAuthorIndex) FindAuthor(digests sc.Sequence[primitives.DigestPreRuntime]) (sc.Option[primitives.AccountId], error) {
	i, err := f.authorFinder.FindAuthor(digests)
	if err != nil {
		return sc.Option[primitives.AccountId]{}, err
	}
	if !i.HasValue {
		return sc.Option[primitives.AccountId]{}, errAuthorNotFound
	}

	validators, err := f.sessionModule.Validators()
	if err != nil {
		return sc.Option[primitives.AccountId]{}, err
	}

	if i.Value >= sc.U32(len(validators)) {
		return sc.Option[primitives.AccountId]{}, errAuthorIndexOutOfBounds
	}

	return sc.NewOption[primitives.AccountId](validators[i.Value]), nil
}
