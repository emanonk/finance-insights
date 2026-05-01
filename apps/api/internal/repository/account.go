package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

// AccountRepository provides read access to account data.
type AccountRepository struct {
	pool *pgxpool.Pool
}

// NewAccountRepository returns an AccountRepository bound to the given pool.
func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

// List returns all known accounts ordered by bank name and account number.
func (r *AccountRepository) List(ctx context.Context) ([]domain.Account, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, bank_name, account_number
		FROM accounts
		ORDER BY bank_name, account_number
	`)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	var out []domain.Account
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(&a.ID, &a.BankName, &a.AccountNumber); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// FindOrCreate returns the id of the account matching bankName and accountNumber,
// creating a new account (with a generated id) if none exists yet.
func (r *AccountRepository) FindOrCreate(ctx context.Context, bankName, accountNumber string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO accounts (id, bank_name, account_number)
		VALUES (gen_random_uuid()::text, $1, $2)
		ON CONFLICT (bank_name, account_number) DO UPDATE SET bank_name = EXCLUDED.bank_name
		RETURNING id
	`, bankName, accountNumber).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("find or create account: %w", err)
	}
	return id, nil
}
