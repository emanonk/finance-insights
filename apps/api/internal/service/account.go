package service

import (
	"context"
	"fmt"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

type accountRepo interface {
	List(ctx context.Context) ([]domain.Account, error)
}

// Account serves account read queries.
type Account struct {
	repo accountRepo
}

// NewAccount constructs an Account service.
func NewAccount(repo accountRepo) *Account {
	return &Account{repo: repo}
}

// List returns all known accounts.
func (s *Account) List(ctx context.Context) ([]domain.Account, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return items, nil
}
