package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AccountRepository provides read access to account data.
type AccountRepository struct {
	pool *pgxpool.Pool
}

// NewAccountRepository returns an AccountRepository bound to the given pool.
func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

// BankName returns the bank_name for the given account id.
func (r *AccountRepository) BankName(ctx context.Context, accountID int64) (string, error) {
	var name string
	if err := r.pool.QueryRow(ctx,
		`SELECT bank_name FROM accounts WHERE id = $1`, accountID,
	).Scan(&name); err != nil {
		return "", fmt.Errorf("get account bank name: %w", err)
	}
	return name, nil
}
