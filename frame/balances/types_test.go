package balances

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	issuanceBalance = sc.NewU128(123)
)

var (
	mockStorageTotalIssuance *mocks.StorageValue[sc.U128]
)

func setupNegativeImbalance() negativeImbalance {
	mockStorageTotalIssuance = new(mocks.StorageValue[sc.U128])

	return newNegativeImbalance(issuanceBalance, mockStorageTotalIssuance)
}

func setupPositiveImbalance() positiveImbalance {
	mockStorageTotalIssuance = new(mocks.StorageValue[sc.U128])

	return newPositiveImbalance(issuanceBalance, mockStorageTotalIssuance)
}

func Test_NegativeImbalance_New(t *testing.T) {
	target := setupNegativeImbalance()

	assert.Equal(t, negativeImbalance{issuanceBalance, mockStorageTotalIssuance}, target)
}

func Test_NegativeImbalance_Drop(t *testing.T) {
	target := setupNegativeImbalance()

	mockStorageTotalIssuance.On("Get").Return(sc.NewU128(5), nil)
	mockStorageTotalIssuance.On("Put", sc.NewU128(0)).Return()

	target.Drop()

	mockStorageTotalIssuance.AssertCalled(t, "Get")
	mockStorageTotalIssuance.AssertCalled(t, "Put", sc.NewU128(0))
}

func Test_PositiveImbalance_New(t *testing.T) {
	target := setupPositiveImbalance()

	assert.Equal(t, positiveImbalance{issuanceBalance, mockStorageTotalIssuance}, target)
}

func Test_PositiveImbalance_Drop(t *testing.T) {
	target := setupPositiveImbalance()

	mockStorageTotalIssuance.On("Get").Return(sc.NewU128(5), nil)
	mockStorageTotalIssuance.On("Put", sc.NewU128(128)).Return()

	target.Drop()

	mockStorageTotalIssuance.AssertCalled(t, "Get")
	mockStorageTotalIssuance.AssertCalled(t, "Put", sc.NewU128(128))
}
