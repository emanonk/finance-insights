package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manoskammas/finance-insights/apps/api/internal/parser"
	"github.com/manoskammas/finance-insights/apps/api/internal/repository"
)

// statementWriter persists a single statement row inside the given tx.
type statementWriter interface {
	Insert(ctx context.Context, tx pgx.Tx, s repository.Statement) error
}

// transactionWriter bulk-persists transactions inside the given tx.
type transactionWriter interface {
	InsertBatch(ctx context.Context, tx pgx.Tx, statementID uuid.UUID, txs []repository.Transaction) error
}

// pdfParser parses a PDF file on disk into ParsedTransactions.
type pdfParser interface {
	Parse(pdfPath string) ([]parser.ParsedTransaction, error)
}

// Statement orchestrates statement ingest: save the upload, parse it,
// and persist the statement and its transactions atomically.
type Statement struct {
	pool       *pgxpool.Pool
	parser     pdfParser
	statements statementWriter
	txs        transactionWriter
	storageDir string
}

// NewStatement constructs a Statement service.
func NewStatement(
	pool *pgxpool.Pool,
	p pdfParser,
	statements statementWriter,
	txs transactionWriter,
	storageDir string,
) *Statement {
	return &Statement{
		pool:       pool,
		parser:     p,
		statements: statements,
		txs:        txs,
		storageDir: storageDir,
	}
}

// IngestResult is returned to clients after a successful upload.
type IngestResult struct {
	StatementID      uuid.UUID
	FileName         string
	TransactionCount int
}

// Ingest saves the uploaded PDF to disk, parses it, and persists the
// statement along with all parsed transactions in a single DB transaction.
func (s *Statement) Ingest(ctx context.Context, fileName string, r io.Reader) (IngestResult, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return IngestResult{}, fmt.Errorf("generate id: %w", err)
	}

	storedPath, err := s.saveToDisk(id, r)
	if err != nil {
		return IngestResult{}, err
	}

	parsed, err := s.parser.Parse(storedPath)
	if err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, fmt.Errorf("parse statement: %w", err)
	}

	domainTxs, err := toDomainTransactions(id, parsed)
	if err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, fmt.Errorf("normalize transactions: %w", err)
	}

	stmt := repository.Statement{
		ID:         id,
		FileName:   fileName,
		StoredPath: storedPath,
		UploadedAt: time.Now().UTC(),
	}

	if err := s.persist(ctx, stmt, domainTxs); err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, err
	}

	return IngestResult{
		StatementID:      id,
		FileName:         fileName,
		TransactionCount: len(domainTxs),
	}, nil
}

func (s *Statement) saveToDisk(id uuid.UUID, r io.Reader) (string, error) {
	dir := filepath.Join(s.storageDir, "statements")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create storage dir: %w", err)
	}

	path := filepath.Join(dir, id.String()+".pdf")
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create storage file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("write storage file: %w", err)
	}
	return path, nil
}

func (s *Statement) persist(ctx context.Context, stmt repository.Statement, txs []repository.Transaction) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.statements.Insert(ctx, tx, stmt); err != nil {
		return err
	}
	if err := s.txs.InsertBatch(ctx, tx, stmt.ID, txs); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// toDomainTransactions converts parser output into repository-layer rows.
// Transactions that lack a required field (date, direction, amount, description)
// are skipped rather than failing the whole upload; malformed values fail fast.
func toDomainTransactions(statementID uuid.UUID, parsed []parser.ParsedTransaction) ([]repository.Transaction, error) {
	out := make([]repository.Transaction, 0, len(parsed))
	for i, p := range parsed {
		if strings.TrimSpace(p.Date) == "" ||
			strings.TrimSpace(p.Direction) == "" ||
			strings.TrimSpace(p.Amount) == "" ||
			strings.TrimSpace(p.Description) == "" {
			continue
		}

		date, err := parseDate(p.Date)
		if err != nil {
			return nil, fmt.Errorf("transaction #%d date %q: %w", i, p.Date, err)
		}
		amount, err := normalizeAmountString(p.Amount)
		if err != nil {
			return nil, fmt.Errorf("transaction #%d amount %q: %w", i, p.Amount, err)
		}

		id, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("generate transaction id: %w", err)
		}

		t := repository.Transaction{
			ID:                  id,
			StatementID:         statementID,
			AccountID:           nullableString(p.AccountID),
			Date:                date,
			MerchantIdentifier:  nullableString(p.MerchantIdentifier),
			Description:         p.Description,
			Direction:           p.Direction,
			Amount:              amount,
			MCCCode:             nullableString(p.MCCCode),
			CardMasked:          nullableString(p.CardMasked),
			Reference:           nullableString(p.Reference),
			BankReferenceNumber: nullableString(p.BankReferenceNumber),
			PaymentMethod:       nullableString(p.PaymentMethod),
		}

		if balance, ok, err := optionalAmount(p.BalanceAfterTransaction); err != nil {
			return nil, fmt.Errorf("transaction #%d balance %q: %w", i, p.BalanceAfterTransaction, err)
		} else if ok {
			t.BalanceAfterTransaction = &balance
		}

		out = append(out, t)
	}
	return out, nil
}

// parseDate accepts dd/mm/yyyy (parser output) or ISO YYYY-MM-DD.
func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"02/01/2006", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, errors.New("unrecognized date format")
}

// normalizeAmountString ensures the parser-provided amount is a valid decimal
// number expressible as numeric(14,2). It returns the canonical dot-decimal form.
func normalizeAmountString(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("empty amount")
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return "", fmt.Errorf("not a number: %w", err)
	}
	return s, nil
}

func optionalAmount(s string) (string, bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false, nil
	}
	a, err := normalizeAmountString(s)
	if err != nil {
		return "", false, err
	}
	return a, true, nil
}

func nullableString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}
