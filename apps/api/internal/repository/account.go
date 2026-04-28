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

// FindByBankAndNumber returns the account id matching the given bank name and
// account number as stored in the accounts table.
func (r *AccountRepository) FindByBankAndNumber(ctx context.Context, bankName, accountNumber string) (int64, error) {
	var id int64
	if err := r.pool.QueryRow(ctx,
		`SELECT id FROM accounts WHERE bank_name = $1 AND account_number = $2`,
		bankName, accountNumber,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("find account by bank and number: %w", err)
	}
	return id, nil
}
