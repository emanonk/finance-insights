package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

// TransactionRepository persists and queries transactions.
type TransactionRepository struct {
	pool *pgxpool.Pool
}

// NewTransactionRepository returns a TransactionRepository bound to the given pool.
func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

// InsertBatch bulk-inserts the given transactions inside the supplied pg transaction.
// Duplicate rows (matched by the transactions_dedup_idx unique index) are silently ignored.
// The database generates the id for each row automatically.
func (r *TransactionRepository) InsertBatch(
	ctx context.Context,
	tx pgx.Tx,
	txs []domain.Transaction,
) error {
	if len(txs) == 0 {
		return nil
	}

	const query = `
		INSERT INTO transactions (
			account_id, date, bank_reference, transaction_reference,
			merchant_identifier, balance_before, balance_after,
			amount, direction, raw_data, statement_file_name
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT ON CONSTRAINT transactions_dedup_idx DO NOTHING`

	batch := &pgx.Batch{}
	for _, t := range txs {
		batch.Queue(query,
			t.AccountID,
			t.Date,
			t.BankReference,
			t.TransactionReference,
			t.MerchantIdentifier,
			t.BalanceBefore,
			t.BalanceAfter,
			t.Amount,
			t.Direction,
			t.RawData,
			t.StatementFileName,
		)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for range txs {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}
	}
	return nil
}

// List returns transactions ordered by date desc (then id desc) with pagination,
// along with the total number of rows matching the filter.
// When accountIDs is non-empty, only transactions belonging to those accounts are returned.
func (r *TransactionRepository) List(ctx context.Context, limit, offset int, accountIDs []string) ([]domain.Transaction, int, error) {
	var (
		query string
		args  []any
	)

	if len(accountIDs) > 0 {
		query = `
			SELECT
				id, account_id, date, bank_reference, transaction_reference,
				merchant_identifier, balance_before, balance_after, amount,
				direction, raw_data, statement_file_name,
				COUNT(*) OVER() AS total
			FROM transactions
			WHERE account_id = ANY($3)
			ORDER BY date DESC, id DESC
			LIMIT $1 OFFSET $2`
		args = []any{limit, offset, accountIDs}
	} else {
		query = `
			SELECT
				id, account_id, date, bank_reference, transaction_reference,
				merchant_identifier, balance_before, balance_after, amount,
				direction, raw_data, statement_file_name,
				COUNT(*) OVER() AS total
			FROM transactions
			ORDER BY date DESC, id DESC
			LIMIT $1 OFFSET $2`
		args = []any{limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var (
		out   []domain.Transaction
		total int
	)
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(
			&t.ID,
			&t.AccountID,
			&t.Date,
			&t.BankReference,
			&t.TransactionReference,
			&t.MerchantIdentifier,
			&t.BalanceBefore,
			&t.BalanceAfter,
			&t.Amount,
			&t.Direction,
			&t.RawData,
			&t.StatementFileName,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate transactions: %w", err)
	}

	if out == nil {
		var countQuery string
		var countArgs []any
		if len(accountIDs) > 0 {
			countQuery = `SELECT COUNT(*) FROM transactions WHERE account_id = ANY($1)`
			countArgs = []any{accountIDs}
		} else {
			countQuery = `SELECT COUNT(*) FROM transactions`
		}
		if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("count transactions: %w", err)
		}
	}

	return out, total, nil
}
