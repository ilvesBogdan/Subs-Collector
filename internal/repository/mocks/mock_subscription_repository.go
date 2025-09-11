package mocks

import (
	"context"
	"time"

	"subs-collector/internal/model"

	"github.com/stretchr/testify/mock"
)

// SubscriptionRepository — мок интерфейса репозитория
type SubscriptionRepository struct {
	mock.Mock
}

func (m *SubscriptionRepository) Create(ctx context.Context, s *model.Subscription) (int, error) {
	args := m.Called(ctx, s)
	return args.Int(0), args.Error(1)
}

func (m *SubscriptionRepository) GetByID(ctx context.Context, id int) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *SubscriptionRepository) Update(ctx context.Context, id int, s *model.Subscription) error {
	args := m.Called(ctx, id, s)
	return args.Error(0)
}

func (m *SubscriptionRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *SubscriptionRepository) List(ctx context.Context, userID, serviceName string) ([]model.Subscription, error) {
	args := m.Called(ctx, userID, serviceName)
	if v := args.Get(0); v != nil {
		return v.([]model.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *SubscriptionRepository) SumTotal(ctx context.Context, from, to time.Time, userID, serviceName string) (int, error) {
	args := m.Called(ctx, from, to, userID, serviceName)
	return args.Int(0), args.Error(1)
}
