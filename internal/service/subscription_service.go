package service

import (
	"context"
	"time"

	"subs-collector/internal/model"
	"subs-collector/internal/repository"
)

type SubscriptionService interface {
	Create(ctx context.Context, s *model.Subscription) (int, error)
	GetByID(ctx context.Context, id int) (*model.Subscription, error)
	Update(ctx context.Context, id int, s *model.Subscription) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, userID string, serviceName string) ([]model.Subscription, error)
	SumTotal(ctx context.Context, from time.Time, to time.Time, userID string, serviceName string) (int, error)
}

type subscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: repo}
}

func (s *subscriptionService) Create(ctx context.Context, sub *model.Subscription) (int, error) {
	if sub.StartDate.IsZero() {
		sub.StartDate = time.Now().UTC()
	}
	return s.repo.Create(ctx, sub)
}

func (s *subscriptionService) GetByID(ctx context.Context, id int) (*model.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *subscriptionService) Update(ctx context.Context, id int, sub *model.Subscription) error {
	return s.repo.Update(ctx, id, sub)
}

func (s *subscriptionService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *subscriptionService) List(ctx context.Context, userID string, serviceName string) ([]model.Subscription, error) {
	return s.repo.List(ctx, userID, serviceName)
}

// SumTotal нормализует границы периода к первому числу месяца и считает сумму
func (s *subscriptionService) SumTotal(ctx context.Context, from time.Time, to time.Time, userID string, serviceName string) (int, error) {
	if to.Before(from) {
		return 0, nil
	}
	from = time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
	to = time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.UTC)
	return s.repo.SumTotal(ctx, from, to, userID, serviceName)
}
