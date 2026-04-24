package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// StatementRepository persists statements.
type StatementRepository struct{}

// NewStatementRepository returns a StatementRepository.
func NewStatementRepository() *StatementRepository { return &StatementRepository{} }

// Insert writes a single statement row inside the supplied transaction.
func (r *StatementRepository) Insert(ctx context.Context, tx pgx.Tx, s Statement) error {
	_, err := tx.Exec(ctx, `
        INSERT INTO statements (id, file_name, stored_path, uploaded_at)
        VALUES ($1, $2, $3, $4)
    `, s.ID, s.FileName, s.StoredPath, s.UploadedAt)
	if err != nil {
		return fmt.Errorf("insert statement: %w", err)
	}
	return nil
}
