package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

type fakeTransactionService struct {
	gotLimit  int
	gotOffset int
	result    service.ListResult
	err       error
}

func (f *fakeTransactionService) List(_ context.Context, limit, offset int) (service.ListResult, error) {
	f.gotLimit = limit
	f.gotOffset = offset
	if f.err != nil {
		return service.ListResult{}, f.err
	}
	return f.result, nil
}

func TestTransactionHandler_List_DefaultsWhenNoQuery(t *testing.T) {
	t.Parallel()

	svc := &fakeTransactionService{result: service.ListResult{Items: nil, Total: 0, Limit: 50, Offset: 0}}
	h := &Transaction{Service: svc}

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body transactionListResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Limit != 50 || body.Offset != 0 {
		t.Errorf("limit/offset = %d/%d, want 50/0", body.Limit, body.Offset)
	}
	if body.Items == nil {
		t.Error("items should be empty array, not null")
	}
}

func TestTransactionHandler_List_PassesQueryToService(t *testing.T) {
	t.Parallel()

	svc := &fakeTransactionService{result: service.ListResult{Limit: 25, Offset: 10}}
	h := &Transaction{Service: svc}

	req := httptest.NewRequest(http.MethodGet, "/transactions?limit=25&offset=10", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.gotLimit != 25 || svc.gotOffset != 10 {
		t.Errorf("service got limit=%d offset=%d, want 25/10", svc.gotLimit, svc.gotOffset)
	}
}

func TestTransactionHandler_List_BadQueryReturns400(t *testing.T) {
	t.Parallel()

	svc := &fakeTransactionService{}
	h := &Transaction{Service: svc}

	req := httptest.NewRequest(http.MethodGet, "/transactions?limit=abc", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestTransactionHandler_List_SerializesDTO(t *testing.T) {
	t.Parallel()

	merchant := "STARBUCKS"
	svc := &fakeTransactionService{
		result: service.ListResult{
			Items: []domain.Transaction{{
				ID:                 42,
				AccountID:          "1",
				Date:               time.Date(2024, 7, 8, 0, 0, 0, 0, time.UTC),
				Direction:          "debit",
				Amount:             7149,
				MerchantIdentifier: &merchant,
			}},
			Total:  1,
			Limit:  50,
			Offset: 0,
		},
	}
	h := &Transaction{Service: svc}

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	var body transactionListResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(body.Items))
	}
	item := body.Items[0]
	if item.ID != "42" {
		t.Errorf("id = %q, want %q", item.ID, "42")
	}
	if item.Date != "2024-07-08" {
		t.Errorf("date = %q, want 2024-07-08", item.Date)
	}
	if item.Amount != 7149 {
		t.Errorf("amount = %d, want 7149", item.Amount)
	}
	if item.AccountID != "1" {
		t.Errorf("accountId = %q, want %q", item.AccountID, "1")
	}
}

func TestTransactionHandler_List_ServiceErrorReturns500(t *testing.T) {
	t.Parallel()

	svc := &fakeTransactionService{err: errors.New("db down")}
	h := &Transaction{Service: svc}

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
