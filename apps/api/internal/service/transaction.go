package service

import (
	"context"
	"fmt"

	"github.com/manoskammas/finance-insights/apps/api/internal/repository"
)

const (
	defaultTransactionLimit = 50
	maxTransactionLimit     = 200
)

// transactionReader is the subset of the transaction repository used by the service.
type transactionReader interface {
	List(ctx context.Context, limit, offset int) ([]repository.Transaction, int, error)
}

// Transaction serves read-side transaction queries.
type Transaction struct {
	repo transactionReader
}

// NewTransaction constructs a Transaction service.
func NewTransaction(repo transactionReader) *Transaction {
	return &Transaction{repo: repo}
}

// ListResult is the paginated result of a transaction listing.
type ListResult struct {
	Items  []repository.Transaction
	Total  int
	Limit  int
	Offset int
}

// List returns the page of transactions indicated by limit/offset.
// Callers may pass 0 or negative values; they are clamped to safe defaults.
func (s *Transaction) List(ctx context.Context, limit, offset int) (ListResult, error) {
	limit = clampLimit(limit)
	if offset < 0 {
		offset = 0
	}

	items, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return ListResult{}, fmt.Errorf("list transactions: %w", err)
	}
	return ListResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultTransactionLimit
	}
	if limit > maxTransactionLimit {
		return maxTransactionLimit
	}
	return limit
}
