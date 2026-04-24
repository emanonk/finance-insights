package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
// All rows are associated with the provided statementID.
func (r *TransactionRepository) InsertBatch(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
	txs []Transaction,
) error {
	if len(txs) == 0 {
		return nil
	}

	columns := []string{
		"id", "statement_id", "account_id", "date", "merchant_identifier",
		"description", "direction", "amount", "balance_after_transaction",
		"mcc_code", "card_masked", "reference", "bank_reference_number",
		"payment_method",
	}

	rows := make([][]any, len(txs))
	for i, t := range txs {
		rows[i] = []any{
			t.ID,
			statementID,
			t.AccountID,
			t.Date,
			t.MerchantIdentifier,
			t.Description,
			t.Direction,
			t.Amount,
			t.BalanceAfterTransaction,
			t.MCCCode,
			t.CardMasked,
			t.Reference,
			t.BankReferenceNumber,
			t.PaymentMethod,
		}
	}

	_, err := tx.CopyFrom(ctx, pgx.Identifier{"transactions"}, columns, pgx.CopyFromRows(rows))
	if err != nil {
		return fmt.Errorf("copy transactions: %w", err)
	}
	return nil
}

// List returns transactions ordered by date desc (then id desc) with pagination,
// along with the total number of rows in the table.
func (r *TransactionRepository) List(ctx context.Context, limit, offset int) ([]Transaction, int, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT
            id,
            statement_id,
            account_id,
            date,
            merchant_identifier,
            description,
            direction,
            amount::text,
            balance_after_transaction::text,
            mcc_code,
            card_masked,
            reference,
            bank_reference_number,
            payment_method,
            created_at,
            COUNT(*) OVER() AS total
        FROM transactions
        ORDER BY date DESC, id DESC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var (
		out   []Transaction
		total int
	)
	for rows.Next() {
		var (
			t       Transaction
			balance *string
		)
		if err := rows.Scan(
			&t.ID,
			&t.StatementID,
			&t.AccountID,
			&t.Date,
			&t.MerchantIdentifier,
			&t.Description,
			&t.Direction,
			&t.Amount,
			&balance,
			&t.MCCCode,
			&t.CardMasked,
			&t.Reference,
			&t.BankReferenceNumber,
			&t.PaymentMethod,
			&t.CreatedAt,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		t.BalanceAfterTransaction = balance
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate transactions: %w", err)
	}

	if out == nil {
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions`).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("count transactions: %w", err)
		}
	}

	return out, total, nil
}
