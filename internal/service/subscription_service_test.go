package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"subs-collector/internal/model"
	rmocks "subs-collector/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestSumTotal_NormalizesDates проверяет, что сервис нормализует границы к первому числу месяца
func TestSumTotal_NormalizesDates(t *testing.T) {
	from := time.Date(2025, 7, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2025, 9, 20, 10, 0, 0, 0, time.UTC)

	m := new(rmocks.SubscriptionRepository)
	m.On("SumTotal", mock.Anything, mock.MatchedBy(func(ti time.Time) bool { return ti.Day() == 1 }), mock.MatchedBy(func(ti time.Time) bool { return ti.Day() == 1 }), "", "").Return(1200, nil)

	s := NewSubscriptionService(m)
	total, err := s.SumTotal(context.Background(), from, to, "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1200, total)
	m.AssertExpectations(t)
}

// TestCreate_SetsStartDateIfZero — если дата не задана, сервис подставляет текущую
func TestCreate_SetsStartDateIfZero(t *testing.T) {
	m := new(rmocks.SubscriptionRepository)
	m.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(1, nil)
	s := NewSubscriptionService(m)
	sub := &model.Subscription{}
	_, _ = s.Create(context.Background(), sub)
	assert.False(t, sub.StartDate.IsZero(), "ожидалось, что StartDate будет установлен")
}

// TestList_ErrorPropagates — ошибка из репозитория пробрасывается наверх
func TestList_ErrorPropagates(t *testing.T) {
	m := new(rmocks.SubscriptionRepository)
	m.On("List", mock.Anything, "", "").Return(nil, errors.New("boom"))
	s := NewSubscriptionService(m)
	_, err := s.List(context.Background(), "", "")
	assert.Error(t, err)
}
