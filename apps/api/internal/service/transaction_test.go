package service

import (
	"context"
	"errors"
	"testing"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

type fakeTransactionRepo struct {
	gotLimit  int
	gotOffset int
	items     []domain.Transaction
	total     int
	err       error
}

func (f *fakeTransactionRepo) List(_ context.Context, limit, offset int) ([]domain.Transaction, int, error) {
	f.gotLimit = limit
	f.gotOffset = offset
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
}

func TestTransactionService_List_ClampsLimitAndOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{"zero limit defaults", 0, 5, 50, 5},
		{"negative limit defaults", -10, 0, 50, 0},
		{"limit above cap", 500, 0, 200, 0},
		{"limit at cap", 200, 0, 200, 0},
		{"normal limit", 25, 10, 25, 10},
		{"negative offset zeroed", 25, -3, 25, 0},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := &fakeTransactionRepo{}
			svc := NewTransaction(repo)

			result, err := svc.List(context.Background(), tc.limit, tc.offset)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.gotLimit != tc.wantLimit {
				t.Errorf("repo.limit = %d, want %d", repo.gotLimit, tc.wantLimit)
			}
			if repo.gotOffset != tc.wantOffset {
				t.Errorf("repo.offset = %d, want %d", repo.gotOffset, tc.wantOffset)
			}
			if result.Limit != tc.wantLimit {
				t.Errorf("result.Limit = %d, want %d", result.Limit, tc.wantLimit)
			}
			if result.Offset != tc.wantOffset {
				t.Errorf("result.Offset = %d, want %d", result.Offset, tc.wantOffset)
			}
		})
	}
}

func TestTransactionService_List_ReturnsItems(t *testing.T) {
	t.Parallel()

	desc := "one"
	repo := &fakeTransactionRepo{
		items: []domain.Transaction{{ID: 1, Description: &desc, Direction: "debit", Amount: "10.00"}},
		total: 42,
	}
	svc := NewTransaction(repo)

	result, err := svc.List(context.Background(), 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(result.Items))
	}
	if result.Total != 42 {
		t.Errorf("total = %d, want 42", result.Total)
	}
}

func TestTransactionService_List_PropagatesError(t *testing.T) {
	t.Parallel()

	repo := &fakeTransactionRepo{err: errors.New("boom")}
	svc := NewTransaction(repo)

	if _, err := svc.List(context.Background(), 0, 0); err == nil {
		t.Fatal("expected error, got nil")
	}
}
