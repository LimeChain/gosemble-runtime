package mocks

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type AuthorshipEventHandler struct {
	mock.Mock
}

func (a *AuthorshipEventHandler) NoteAuthor(author primitives.AccountId) {
	a.Called(author)
}
