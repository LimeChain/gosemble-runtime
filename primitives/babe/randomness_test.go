package babe

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	randomness = sc.FixedSequence[sc.U8]{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0,
	}
)

func Test_NewRandomness(t *testing.T) {
	assert.Equal(t, randomness, NewRandomness())
}
