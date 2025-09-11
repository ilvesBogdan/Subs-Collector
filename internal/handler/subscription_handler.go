package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"subs-collector/internal/logger"
	"subs-collector/internal/model"
	"subs-collector/internal/service"

	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service service.SubscriptionService
	log     *logger.Logger
}

func NewSubscriptionHandler(s service.SubscriptionService, l *logger.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{service: s, log: l}
}

func (h *SubscriptionHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/subscriptions", h.handleListOrCreate)
	mux.HandleFunc("/subscriptions/", h.handleByID)
	mux.HandleFunc("/subscriptions/summary", h.handleSummary)
}

type subscriptionDTO struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
}

func (h *SubscriptionHandler) handleListOrCreate(w http.ResponseWriter, r *http.Request) {
	h.log.Info("incoming request", "method", r.Method, "path", r.URL.Path)
	if r.Method == http.MethodPost {
		h.create(w, r)
		return
	}

	if r.Method == http.MethodGet {
		h.list(w, r)
		return
	}

	h.respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (h *SubscriptionHandler) handleByID(w http.ResponseWriter, r *http.Request) {
	h.log.Info("incoming request", "method", r.Method, "path", r.URL.Path)
	idStr := r.URL.Path[len("/subscriptions/"):]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		h.log.Error("invalid id", "id", idStr, "err", err)
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		h.respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *SubscriptionHandler) create(w http.ResponseWriter, r *http.Request) {
	var dto subscriptionDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.log.Error("decode body error", "err", err)
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if _, err := uuid.Parse(dto.UserID); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
		return
	}
	start, err := parseData(dto.StartDate)

	if err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid start_date"})
		return
	}

	sub := model.Subscription{
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		UserID:      dto.UserID,
		StartDate:   start,
	}

	id, err := h.service.Create(r.Context(), &sub)
	if err != nil {
		h.log.Error("create error", "err", err)
		h.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *SubscriptionHandler) get(w http.ResponseWriter, r *http.Request, id int) {
	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.log.Error("get error", "id", id, "err", err)
		h.respondJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	h.respondJSON(w, http.StatusOK, sub)
}

func (h *SubscriptionHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var dto subscriptionDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.log.Error("decode body error", "err", err)
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if _, err := uuid.Parse(dto.UserID); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
		return
	}
	start, err := parseData(dto.StartDate)

	if err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid start_date"})
		return
	}

	sub := model.Subscription{
		ID:          id,
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		UserID:      dto.UserID,
		StartDate:   start,
	}

	if err := h.service.Update(r.Context(), id, &sub); err != nil {
		h.log.Error("update error", "id", id, "err", err)
		h.respondJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *SubscriptionHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.log.Error("delete error", "id", id, "err", err)
		h.respondJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *SubscriptionHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := q.Get("user_id")
	serviceName := q.Get("service_name")
	items, err := h.service.List(r.Context(), userID, serviceName)

	if err != nil {
		h.log.Error("list error", "err", err)
		h.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	h.respondJSON(w, http.StatusOK, items)
}

func (h *SubscriptionHandler) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")
	userID := q.Get("user_id")
	serviceName := q.Get("service_name")

	from, err := parseData(fromStr)
	if err != nil {
		h.log.Error("invalid from", "from", fromStr, "err", err)
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid from"})
		return
	}
	to, err := parseData(toStr)
	if err != nil {
		h.log.Error("invalid to", "to", toStr, "err", err)
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid to"})
		return
	}

	total, err := h.service.SumTotal(r.Context(), from, to, userID, serviceName)
	if err != nil {
		h.log.Error("summary error", "err", err)
		h.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]int{"total": total})
}

func parseData(s string) (time.Time, error) {
	if len(s) != 7 || s[2] != '-' {
		return time.Time{}, fmt.Errorf("bad format, expected MM-YYYY")
	}
	month, err := strconv.Atoi(s[0:2])
	if err != nil || month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("bad month")
	}
	year, err := strconv.Atoi(s[3:7])
	if err != nil || year < 1900 || year > 3000 {
		return time.Time{}, fmt.Errorf("bad year")
	}
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
}

func (h *SubscriptionHandler) respondJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
