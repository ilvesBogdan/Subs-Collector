package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"subs-collector/internal/logger"
	"subs-collector/internal/model"
)

type fakeService struct {
	createdID  int
	createdErr error
}

func (f *fakeService) Create(_ context.Context, _ *model.Subscription) (int, error) {
	return f.createdID, f.createdErr
}
func (f *fakeService) GetByID(_ context.Context, _ int) (*model.Subscription, error) {
	return &model.Subscription{ID: 1, ServiceName: "S", Price: 100, UserID: "00000000-0000-0000-0000-000000000000", StartDate: time.Now()}, nil
}
func (f *fakeService) Update(_ context.Context, _ int, _ *model.Subscription) error { return nil }
func (f *fakeService) Delete(_ context.Context, _ int) error                        { return nil }
func (f *fakeService) List(_ context.Context, _ string, _ string) ([]model.Subscription, error) {
	return []model.Subscription{}, nil
}
func (f *fakeService) SumTotal(_ context.Context, _ time.Time, _ time.Time, _ string, _ string) (int, error) {
	return 0, nil
}

func TestCreate_ValidBody(t *testing.T) {
	l := logger.New()
	s := &fakeService{createdID: 42}
	h := NewSubscriptionHandler(s, l)

	body := map[string]interface{}{
		"service_name": "Netflix",
		"price":        999,
		"user_id":      "00000000-0000-0000-0000-000000000000",
		"start_date":   "07-2025",
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.handleListOrCreate(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("ожидался 201, получил %d", rec.Code)
	}
}
