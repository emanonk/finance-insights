package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
	"github.com/manoskammas/finance-insights/apps/api/internal/parsers"
)

// transactionWriter bulk-persists transactions inside the given tx.
type transactionWriter interface {
	InsertBatch(ctx context.Context, tx pgx.Tx, txs []domain.Transaction) error
}

// statementParser parses a PDF file on disk into a ParseResult.
type statementParser interface {
	Parse(bankName, pdfPath string) (parsers.ParseResult, error)
}

// accountResolver finds or creates an account for the given bank name and
// account number, returning the account id.
type accountResolver interface {
	FindOrCreate(ctx context.Context, bankName, accountNumber string) (string, error)
}

// Statement orchestrates statement ingest: save the upload, parse it,
// and persist the transactions atomically.
type Statement struct {
	pool       *pgxpool.Pool
	parser     statementParser
	txs        transactionWriter
	accounts   accountResolver
	storageDir string
}

// NewStatement constructs a Statement service.
func NewStatement(
	pool *pgxpool.Pool,
	p statementParser,
	txs transactionWriter,
	accounts accountResolver,
	storageDir string,
) *Statement {
	return &Statement{
		pool:       pool,
		parser:     p,
		txs:        txs,
		accounts:   accounts,
		storageDir: storageDir,
	}
}

// IngestResult is returned to clients after a successful upload.
type IngestResult struct {
	FileName         string
	TransactionCount int
}

// Ingest saves the uploaded PDF to disk, parses it with the bank's registered
// parsers, resolves the account from the account number found in the statement,
// and persists all parsed transactions in a single DB transaction.
func (s *Statement) Ingest(ctx context.Context, bankName string, fileName string, r io.Reader) (IngestResult, error) {
	storedPath, err := s.saveToDisk(fileName, r)
	if err != nil {
		return IngestResult{}, err
	}

	result, err := s.parser.Parse(bankName, storedPath)
	if err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, fmt.Errorf("parse statement: %w", err)
	}

	accountID, err := s.accounts.FindOrCreate(ctx, bankName, result.AccountNumber)
	if err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, fmt.Errorf("resolve account: %w", err)
	}

	domainTxs, err := toDomainTransactions(accountID, fileName, result.Transactions)
	if err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, fmt.Errorf("normalize transactions: %w", err)
	}

	if err := s.persist(ctx, domainTxs); err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, err
	}

	return IngestResult{
		FileName:         fileName,
		TransactionCount: len(domainTxs),
	}, nil
}

func (s *Statement) saveToDisk(fileName string, r io.Reader) (string, error) {
	dir := filepath.Join(s.storageDir, "statements")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create storage dir: %w", err)
	}

	safeName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fileName))
	path := filepath.Join(dir, safeName)
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

func (s *Statement) persist(ctx context.Context, txs []domain.Transaction) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.txs.InsertBatch(ctx, tx, txs); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// toDomainTransactions converts parser output into domain transactions.
// Transactions missing a date, direction, or non-zero amount are skipped.
// accountID is the resolved DB account id and overrides whatever the parser set.
func toDomainTransactions(accountID, fileName string, parsed []parsers.ParsedTransaction) ([]domain.Transaction, error) {
	out := make([]domain.Transaction, 0, len(parsed))
	for _, p := range parsed {
		if p.Date.IsZero() || strings.TrimSpace(p.Direction) == "" || p.Amount == 0 {
			continue
		}
		t := domain.Transaction{
			AccountID:            accountID,
			Date:                 p.Date,
			BankReference:        p.BankReference,
			TransactionReference: p.TransactionReference,
			MerchantIdentifier:   p.MerchantIdentifier,
			BalanceBefore:        p.BalanceBefore,
			BalanceAfter:         p.BalanceAfter,
			Amount:               p.Amount,
			Direction:            strings.ToLower(p.Direction),
			RawData:              p.RawData,
			StatementFileName:    &fileName,
		}
		out = append(out, t)
	}
	return out, nil
}
