package service

import (
	"context"
	"fmt"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

const defaultTopMerchantsLimit = 30

type merchantStore interface {
	TopIdentifiers(ctx context.Context, limit int) ([]domain.IdentifierCount, error)
	Upsert(ctx context.Context, identifierName, primaryTagName string, secondaryTagNames []string) (*domain.Merchant, error)
}

// Merchant serves merchant/tag read and write operations.
type Merchant struct {
	repo merchantStore
}

// NewMerchant constructs a Merchant service.
func NewMerchant(repo merchantStore) *Merchant {
	return &Merchant{repo: repo}
}

// TopIdentifiers returns the most frequent merchant identifiers, capped at 100.
func (s *Merchant) TopIdentifiers(ctx context.Context, limit int) ([]domain.IdentifierCount, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultTopMerchantsLimit
	}
	items, err := s.repo.TopIdentifiers(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("top identifiers: %w", err)
	}
	return items, nil
}

// UpsertMerchant creates or updates a merchant record with the given tags.
func (s *Merchant) UpsertMerchant(
	ctx context.Context,
	identifierName string,
	primaryTagName string,
	secondaryTagNames []string,
) (*domain.Merchant, error) {
	if identifierName == "" {
		return nil, fmt.Errorf("identifierName is required")
	}
	if primaryTagName == "" {
		return nil, fmt.Errorf("primaryTagName is required")
	}
	m, err := s.repo.Upsert(ctx, identifierName, primaryTagName, secondaryTagNames)
	if err != nil {
		return nil, fmt.Errorf("upsert merchant: %w", err)
	}
	return m, nil
}
