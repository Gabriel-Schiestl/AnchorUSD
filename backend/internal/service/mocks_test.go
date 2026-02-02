package service

import (
	"math/big"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockPriceFeedAPI struct {
	mock.Mock
}

func (m *MockPriceFeedAPI) GetBtcUsdPrice() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockPriceFeedAPI) GetEthUsdPrice() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type MockCacheStore struct {
	mock.Mock
}

func (m *MockCacheStore) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheStore) Set(key string, value any, expiration time.Duration) (string, error) {
	args := m.Called(key, value, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockCacheStore) Add(key string, amountInWei *big.Int) (*big.Int, error) {
	args := m.Called(key, amountInWei)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockCacheStore) MultiAdd(keys []string, amountInWei *big.Int) error {
	args := m.Called(keys, amountInWei)
	return args.Error(0)
}

func (m *MockCacheStore) HGetAll(key string) (map[string]string, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockCacheStore) HGet(key, field string) (string, error) {
	args := m.Called(key, field)
	return args.String(0), args.Error(1)
}

func (m *MockCacheStore) HSet(key, field string, value any) error {
	args := m.Called(key, field, value)
	return args.Error(0)
}

func (m *MockCacheStore) HAdd(key, field string, amountInWei *big.Int) (*big.Int, error) {
	args := m.Called(key, field, amountInWei)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockCacheStore) SSet(key string, members ...string) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *MockCacheStore) SGetAll(key string) ([]string, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheStore) FlushAll() error {
	args := m.Called()
	return args.Error(0)
}
